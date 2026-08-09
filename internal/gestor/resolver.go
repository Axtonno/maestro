package gestor

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgGestor.Resolver = (*Resolver)(nil)

type dependencyGraphView interface {
	State() (generation uint64, current bool)
	DependencyPlan(
		pkgRuntime.ComponentID,
	) (generation uint64, dependencies []pkgRuntime.ComponentID, err error)
}

type Resolver struct {
	registry *Registry
	graph    dependencyGraphView
}

type resolutionEvaluation struct {
	metadata        pkgGestor.SnapshotMetadata
	candidates      []pkgGestor.Descriptor
	dependencyPlans map[pkgGestor.Target][]pkgGestor.Target
	graphGeneration uint64
	graphUsed       bool
}

func NewResolver(
	registry *Registry,
	graph dependencyGraphView,
) (*Resolver, error) {
	if registry == nil {
		return nil, fmt.Errorf("resolver registry is nil: %w", pkgGestor.ErrInvalidResolution)
	}
	if nilInterface(graph) {
		return nil, fmt.Errorf("resolver dependency graph view is nil: %w", pkgGestor.ErrInvalidResolution)
	}

	return &Resolver{registry: registry, graph: graph}, nil
}

func (resolver *Resolver) Candidates(
	query pkgGestor.Query,
) ([]pkgGestor.Descriptor, error) {
	evaluation, err := resolver.evaluate(query)
	if err != nil {
		return nil, err
	}

	return slices.Clone(evaluation.candidates), nil
}

func (resolver *Resolver) Resolve(
	query pkgGestor.Query,
) (pkgGestor.Resolution, error) {
	evaluation, err := resolver.evaluate(query)
	if err != nil {
		return pkgGestor.Resolution{}, err
	}

	selected := pkgGestor.Descriptor{}
	reason := pkgGestor.ResolutionSingleCandidate
	if len(evaluation.candidates) == 1 {
		selected = evaluation.candidates[0]
	} else {
		for _, preferred := range query.PreferredTargets() {
			for _, candidate := range evaluation.candidates {
				if candidate.Target == preferred {
					selected = candidate
					reason = pkgGestor.ResolutionPreferredTarget
					break
				}
			}
			if selected.Capability != "" {
				break
			}
		}
		if selected.Capability == "" {
			if err := resolver.ensureCurrent(evaluation); err != nil {
				return pkgGestor.Resolution{}, err
			}
			return pkgGestor.Resolution{}, fmt.Errorf(
				"resolve capability %q among [%s]: %w",
				query.Capability(),
				formatCandidates(evaluation.candidates),
				pkgGestor.ErrAmbiguous,
			)
		}
	}

	if err := resolver.ensureCurrent(evaluation); err != nil {
		return pkgGestor.Resolution{}, err
	}
	resolution, err := pkgGestor.NewResolution(
		selected,
		evaluation.metadata,
		reason,
		evaluation.dependencyPlans[selected.Target],
	)
	if err != nil {
		return pkgGestor.Resolution{}, fmt.Errorf("construct capability resolution: %w", err)
	}

	return resolution, nil
}

func (resolver *Resolver) evaluate(
	query pkgGestor.Query,
) (resolutionEvaluation, error) {
	if err := query.Validate(); err != nil {
		return resolutionEvaluation{}, err
	}

	snapshot, declared := resolver.registry.resolutionSnapshot(query.Capability())
	metadata := snapshot.Metadata()
	if !metadata.Current {
		return resolutionEvaluation{}, fmt.Errorf(
			"resolve capability %q from generation %d: %w",
			query.Capability(),
			metadata.Generation,
			pkgGestor.ErrStaleSnapshot,
		)
	}
	if len(declared) == 0 {
		return resolutionEvaluation{}, fmt.Errorf(
			"capability %q has no declarations: %w",
			query.Capability(),
			pkgGestor.ErrNotFound,
		)
	}

	matching := make([]pkgGestor.Descriptor, 0, len(declared))
	for _, descriptor := range declared {
		if query.TargetKind() != "" && descriptor.Target.Kind != query.TargetKind() {
			continue
		}
		if query.Scope() != "" && descriptor.Target.Scope != query.Scope() {
			continue
		}
		if query.Model() != "" && descriptor.Target.Model != query.Model() {
			continue
		}
		matching = append(matching, descriptor)
	}
	if len(matching) == 0 {
		return resolutionEvaluation{}, fmt.Errorf(
			"capability %q has no declaration matching target kind %q scope %q model %q: %w",
			query.Capability(),
			query.TargetKind(),
			query.Scope(),
			query.Model(),
			pkgGestor.ErrNotFound,
		)
	}

	available := make([]pkgGestor.Descriptor, 0, len(matching))
	for _, descriptor := range matching {
		if descriptor.Availability == pkgGestor.AvailabilityUnavailable {
			continue
		}
		if query.RequireAvailable() &&
			descriptor.Availability != pkgGestor.AvailabilityAvailable {
			continue
		}
		available = append(available, descriptor)
	}
	if len(available) == 0 {
		return resolutionEvaluation{}, fmt.Errorf(
			"capability %q declarations [%s] are not operationally eligible: %w",
			query.Capability(),
			formatCandidates(matching),
			pkgGestor.ErrUnavailable,
		)
	}

	evaluation := resolutionEvaluation{
		metadata:        metadata,
		candidates:      make([]pkgGestor.Descriptor, 0, len(available)),
		dependencyPlans: make(map[pkgGestor.Target][]pkgGestor.Target),
	}
	usesGraph := false
	for _, descriptor := range available {
		if descriptor.Target.Kind == pkgGestor.TargetKindComponent {
			usesGraph = true
			break
		}
	}
	if usesGraph {
		generation, current := resolver.graph.State()
		if !current {
			return resolutionEvaluation{}, fmt.Errorf(
				"dependency graph generation %d is stale: %w",
				generation,
				pkgGestor.ErrStaleSnapshot,
			)
		}
		evaluation.graphUsed = true
		evaluation.graphGeneration = generation
	}

	for _, descriptor := range available {
		if descriptor.Target.Kind != pkgGestor.TargetKindComponent {
			evaluation.candidates = append(evaluation.candidates, descriptor)
			continue
		}

		generation, dependencies, err := resolver.graph.DependencyPlan(
			pkgRuntime.ComponentID(descriptor.Target.ID),
		)
		if generation != evaluation.graphGeneration {
			return resolutionEvaluation{}, fmt.Errorf(
				"dependency graph changed from generation %d to %d: %w",
				evaluation.graphGeneration,
				generation,
				pkgGestor.ErrStaleSnapshot,
			)
		}
		if err != nil {
			switch {
			case errors.Is(err, pkgRuntime.ErrNotFound):
				continue
			case errors.Is(err, pkgRuntime.ErrInvalidState):
				return resolutionEvaluation{}, fmt.Errorf("read dependency plan for component %q: %w: %w", descriptor.Target.ID, err, pkgGestor.ErrStaleSnapshot)
			default:
				return resolutionEvaluation{}, fmt.Errorf("read dependency plan for component %q: %w: %w", descriptor.Target.ID, err, pkgGestor.ErrUnavailable)
			}
		}
		plan := make([]pkgGestor.Target, 0, len(dependencies))
		for _, dependency := range dependencies {
			plan = append(plan, pkgGestor.Target{
				Kind:  pkgGestor.TargetKindComponent,
				ID:    string(dependency),
				Scope: pkgGestor.ScopeComponent,
			})
		}
		evaluation.dependencyPlans[descriptor.Target] = plan
		evaluation.candidates = append(evaluation.candidates, descriptor)
	}

	if len(evaluation.candidates) == 0 {
		return resolutionEvaluation{}, fmt.Errorf(
			"capability %q candidates [%s] are absent from the dependency graph: %w",
			query.Capability(),
			formatCandidates(available),
			pkgGestor.ErrUnavailable,
		)
	}
	slices.SortFunc(evaluation.candidates, func(left, right pkgGestor.Descriptor) int {
		return left.Compare(right)
	})
	if err := resolver.ensureCurrent(evaluation); err != nil {
		return resolutionEvaluation{}, err
	}

	return evaluation, nil
}

func (resolver *Resolver) ensureCurrent(evaluation resolutionEvaluation) error {
	if !resolver.registry.snapshotIsCurrent(evaluation.metadata.Generation) {
		return fmt.Errorf(
			"capability snapshot generation %d changed during resolution: %w",
			evaluation.metadata.Generation,
			pkgGestor.ErrStaleSnapshot,
		)
	}
	if evaluation.graphUsed {
		generation, current := resolver.graph.State()
		if !current || generation != evaluation.graphGeneration {
			return fmt.Errorf(
				"dependency graph generation %d changed to %d during resolution: %w",
				evaluation.graphGeneration,
				generation,
				pkgGestor.ErrStaleSnapshot,
			)
		}
	}

	return nil
}

func formatCandidates(candidates []pkgGestor.Descriptor) string {
	values := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		value := fmt.Sprintf(
			"%s/%s:%s",
			candidate.Target.Kind,
			candidate.Target.ID,
			candidate.Target.Scope,
		)
		if candidate.Target.Model != "" {
			value += ":" + candidate.Target.Model
		}
		value += "=" + string(candidate.Availability)
		values = append(values, value)
	}

	return strings.Join(values, ", ")
}
