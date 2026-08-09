package benchmark

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInvalidManifest = errors.New("invalid benchmark manifest")
	ErrInvalidReport   = errors.New("invalid benchmark report")
)

var requiredRedactionFields = []string{
	"prompts",
	"responses",
	"tool_arguments",
	"tool_results",
	"credentials",
	"user_paths",
}

func (m Manifest) Validate() error {
	if m.Version != ManifestSchemaVersion {
		return fmt.Errorf(
			"manifest version %d is not supported, expected %d: %w",
			m.Version,
			ManifestSchemaVersion,
			ErrInvalidManifest,
		)
	}
	if strings.TrimSpace(m.Owner) == "" {
		return fmt.Errorf("manifest owner is empty: %w", ErrInvalidManifest)
	}
	if len(m.Providers) == 0 {
		return fmt.Errorf("manifest has no providers: %w", ErrInvalidManifest)
	}
	redactedFields := make(map[string]struct{}, len(m.ReportRedaction.Exclude))
	for _, field := range m.ReportRedaction.Exclude {
		field = strings.TrimSpace(field)
		if field == "" {
			return fmt.Errorf(
				"manifest has an empty redaction field: %w",
				ErrInvalidManifest,
			)
		}
		if _, exists := redactedFields[field]; exists {
			return fmt.Errorf(
				"manifest redaction field %q is duplicated: %w",
				field,
				ErrInvalidManifest,
			)
		}
		redactedFields[field] = struct{}{}
	}
	for _, required := range requiredRedactionFields {
		if _, exists := redactedFields[required]; !exists {
			return fmt.Errorf(
				"manifest redaction field %q is missing: %w",
				required,
				ErrInvalidManifest,
			)
		}
	}
	for providerID := range m.Providers {
		if strings.TrimSpace(providerID) == "" {
			return fmt.Errorf("manifest has an empty provider ID: %w", ErrInvalidManifest)
		}
	}
	if len(m.Scenarios) == 0 {
		return fmt.Errorf("manifest has no scenarios: %w", ErrInvalidManifest)
	}

	seenScenarios := make(map[string]struct{}, len(m.Scenarios))
	for _, scenario := range m.Scenarios {
		if err := scenario.Validate(); err != nil {
			return fmt.Errorf("manifest scenario: %w", err)
		}
		if _, exists := seenScenarios[scenario.ID]; exists {
			return fmt.Errorf(
				"manifest scenario %q is duplicated: %w",
				scenario.ID,
				ErrInvalidManifest,
			)
		}
		seenScenarios[scenario.ID] = struct{}{}
	}

	requiredStates := map[ResultState]bool{
		ResultPassed:      false,
		ResultFailed:      false,
		ResultSkipped:     false,
		ResultUnsupported: false,
	}
	for _, state := range m.ResultStates {
		if !state.Valid() {
			return fmt.Errorf(
				"manifest result state %q is invalid: %w",
				state,
				ErrInvalidManifest,
			)
		}
		if requiredStates[state] {
			return fmt.Errorf(
				"manifest result state %q is duplicated: %w",
				state,
				ErrInvalidManifest,
			)
		}
		requiredStates[state] = true
	}
	for state, present := range requiredStates {
		if !present {
			return fmt.Errorf(
				"manifest result state %q is missing: %w",
				state,
				ErrInvalidManifest,
			)
		}
	}

	return nil
}

func (s ScenarioDefinition) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("scenario ID is empty: %w", ErrInvalidManifest)
	}
	if strings.TrimSpace(s.Capability) == "" {
		return fmt.Errorf(
			"scenario %q capability is empty: %w",
			s.ID,
			ErrInvalidManifest,
		)
	}
	if strings.TrimSpace(s.ModelRole) == "" {
		return fmt.Errorf(
			"scenario %q model role is empty: %w",
			s.ID,
			ErrInvalidManifest,
		)
	}
	if strings.TrimSpace(s.Cleanup) == "" {
		return fmt.Errorf(
			"scenario %q cleanup is empty: %w",
			s.ID,
			ErrInvalidManifest,
		)
	}

	return nil
}

func (s ResultState) Valid() bool {
	switch s {
	case ResultPassed, ResultFailed, ResultSkipped, ResultUnsupported:
		return true
	default:
		return false
	}
}

func (r IterationResult) Validate() error {
	if !r.State.Valid() {
		return fmt.Errorf("iteration state %q is invalid", r.State)
	}
	if r.State == ResultFailed && r.Error == nil {
		return errors.New("failed iteration has no classified error")
	}
	if r.State != ResultFailed && r.Error != nil {
		return fmt.Errorf("%s iteration has a classified error", r.State)
	}
	if err := validateErrorRecord(r.Error); err != nil {
		return err
	}
	if r.State != ResultPassed && len(r.Measurements) != 0 {
		return fmt.Errorf("%s iteration has measurements", r.State)
	}
	if (r.State == ResultSkipped || r.State == ResultUnsupported) &&
		strings.TrimSpace(r.ReasonCode) == "" {
		return fmt.Errorf("%s iteration has no reason code", r.State)
	}
	if (r.State == ResultPassed || r.State == ResultFailed) &&
		strings.TrimSpace(r.ReasonCode) != "" {
		return fmt.Errorf("%s iteration has a reason code", r.State)
	}

	seenMeasurements := make(map[string]struct{}, len(r.Measurements))
	for _, measurement := range r.Measurements {
		if err := measurement.Validate(); err != nil {
			return err
		}
		key := measurement.Name + "\x00" + measurement.Unit + "\x00" +
			measurement.Scope + "\x00" + measurement.Method
		if _, exists := seenMeasurements[key]; exists {
			return fmt.Errorf("measurement %q is duplicated", measurement.Name)
		}
		seenMeasurements[key] = struct{}{}
	}

	return nil
}

func (m Measurement) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("measurement name is empty")
	}
	if strings.TrimSpace(m.Unit) == "" {
		return fmt.Errorf("measurement %q unit is empty", m.Name)
	}
	if math.IsNaN(m.Value) || math.IsInf(m.Value, 0) {
		return fmt.Errorf("measurement %q value is not finite", m.Name)
	}

	return nil
}

func (r Report) Validate() error {
	if r.SchemaVersion != ReportSchemaVersion {
		return fmt.Errorf(
			"report schema %q is not supported, expected %q: %w",
			r.SchemaVersion,
			ReportSchemaVersion,
			ErrInvalidReport,
		)
	}
	if strings.TrimSpace(r.RunID) == "" {
		return fmt.Errorf("report run ID is empty: %w", ErrInvalidReport)
	}
	if r.CreatedAt.IsZero() || r.CompletedAt.IsZero() ||
		r.CompletedAt.Before(r.CreatedAt) {
		return fmt.Errorf("report timestamps are invalid: %w", ErrInvalidReport)
	}
	if r.DurationMS < 0 || math.IsNaN(r.DurationMS) || math.IsInf(r.DurationMS, 0) {
		return fmt.Errorf("report duration is negative: %w", ErrInvalidReport)
	}
	if r.Metadata.ManifestVersion != ManifestSchemaVersion ||
		strings.TrimSpace(r.Metadata.ManifestOwner) == "" {
		return fmt.Errorf("report manifest metadata is invalid: %w", ErrInvalidReport)
	}
	if err := r.Configuration.Validate(); err != nil {
		return fmt.Errorf("report configuration is invalid: %v: %w", err, ErrInvalidReport)
	}
	if len(r.Scenarios) == 0 {
		return fmt.Errorf("report has no scenarios: %w", ErrInvalidReport)
	}
	for _, scenario := range r.Scenarios {
		if err := scenario.Scenario.Validate(); err != nil {
			return fmt.Errorf(
				"report scenario definition %q is invalid: %v: %w",
				scenario.Scenario.ID,
				err,
				ErrInvalidReport,
			)
		}
		if !scenario.State.Valid() || len(scenario.Samples) == 0 {
			return fmt.Errorf(
				"report scenario %q is invalid: %w",
				scenario.Scenario.ID,
				ErrInvalidReport,
			)
		}
		expectedState := ResultPassed
		for _, sample := range scenario.Samples {
			result := IterationResult{
				State: sample.State, ReasonCode: sample.ReasonCode,
				Measurements: sample.Measurements, Error: sample.Error,
			}
			if sample.Iteration.Index < 1 || sample.StartedAt.IsZero() ||
				sample.StartedAt.Before(r.CreatedAt) ||
				sample.StartedAt.After(r.CompletedAt) || sample.DurationMS < 0 ||
				math.IsNaN(sample.DurationMS) || math.IsInf(sample.DurationMS, 0) {
				return fmt.Errorf(
					"report scenario %q has invalid sample metadata: %w",
					scenario.Scenario.ID,
					ErrInvalidReport,
				)
			}
			if err := result.Validate(); err != nil {
				return fmt.Errorf(
					"report scenario %q has invalid sample: %v: %w",
					scenario.Scenario.ID,
					err,
					ErrInvalidReport,
				)
			}
			if err := validateErrorRecord(sample.CleanupError); err != nil {
				return fmt.Errorf(
					"report scenario %q has invalid cleanup error: %v: %w",
					scenario.Scenario.ID,
					err,
					ErrInvalidReport,
				)
			}
			if expectedState == ResultPassed {
				if sample.CleanupError != nil {
					expectedState = ResultFailed
				} else if sample.State != ResultPassed {
					expectedState = sample.State
				}
			}
		}
		if scenario.State != expectedState {
			return fmt.Errorf(
				"report scenario %q state %q does not match samples state %q: %w",
				scenario.Scenario.ID,
				scenario.State,
				expectedState,
				ErrInvalidReport,
			)
		}
		for _, aggregate := range scenario.Aggregates {
			if err := aggregate.Validate(); err != nil {
				return fmt.Errorf(
					"report scenario %q has invalid aggregate: %v: %w",
					scenario.Scenario.ID,
					err,
					ErrInvalidReport,
				)
			}
		}
	}

	return nil
}

func (c ConfigurationProfile) Validate() error {
	if c.Hardware.LogicalCPUs < 0 || c.Hardware.MemoryMB < 0 ||
		c.Hardware.VRAMMB < 0 {
		return errors.New("hardware profile contains a negative value")
	}
	if c.Model.ContextLength < 0 {
		return errors.New("model context length is negative")
	}
	for role, model := range c.Models {
		if strings.TrimSpace(role) == "" {
			return errors.New("model profile role is empty")
		}
		if model.ContextLength < 0 {
			return fmt.Errorf("model profile role %q has a negative context length", role)
		}
	}
	if c.Generation.MaxTokens < 0 || c.Generation.StopCount < 0 {
		return errors.New("generation profile contains a negative count")
	}
	if c.Generation.Temperature != nil &&
		(math.IsNaN(*c.Generation.Temperature) ||
			math.IsInf(*c.Generation.Temperature, 0) ||
			*c.Generation.Temperature < 0) {
		return errors.New("generation temperature is invalid")
	}
	if c.Generation.TopP != nil &&
		(math.IsNaN(*c.Generation.TopP) || math.IsInf(*c.Generation.TopP, 0) ||
			*c.Generation.TopP < 0 || *c.Generation.TopP > 1) {
		return errors.New("generation top_p is invalid")
	}
	if c.Execution.Runs < 1 || c.Execution.Warmup < 0 ||
		c.Execution.TimeoutMS < 0 || c.Execution.CleanupTimeoutMS <= 0 ||
		math.IsNaN(c.Execution.TimeoutMS) || math.IsInf(c.Execution.TimeoutMS, 0) ||
		math.IsNaN(c.Execution.CleanupTimeoutMS) ||
		math.IsInf(c.Execution.CleanupTimeoutMS, 0) {
		return errors.New("execution profile is invalid")
	}
	seenPlugins := make(map[string]struct{}, len(c.Plugins))
	for _, plugin := range c.Plugins {
		if strings.TrimSpace(plugin.ID) == "" {
			return errors.New("plugin profile has an empty ID")
		}
		if _, exists := seenPlugins[plugin.ID]; exists {
			return fmt.Errorf("plugin profile %q is duplicated", plugin.ID)
		}
		seenPlugins[plugin.ID] = struct{}{}
	}

	return nil
}

func (a Aggregate) Validate() error {
	if strings.TrimSpace(a.Name) == "" || strings.TrimSpace(a.Unit) == "" {
		return errors.New("aggregate identity is empty")
	}
	if a.Count < 1 {
		return errors.New("aggregate count is less than one")
	}
	values := []float64{a.Min, a.Median, a.Max}
	if a.P95 != nil {
		values = append(values, *a.P95)
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("aggregate contains a non-finite value")
		}
	}
	if a.Min > a.Median || a.Median > a.Max ||
		(a.P95 != nil && (*a.P95 < a.Median || *a.P95 > a.Max)) {
		return errors.New("aggregate values are not ordered")
	}
	if a.Count < 20 && a.P95 != nil {
		return errors.New("aggregate p95 requires at least 20 samples")
	}

	return nil
}

func validateErrorRecord(record *ErrorRecord) error {
	if record == nil {
		return nil
	}
	if strings.TrimSpace(record.Kind) == "" || strings.TrimSpace(record.Code) == "" {
		return errors.New("classified error identity is empty")
	}
	if record.StatusCode < 0 {
		return errors.New("classified error status code is negative")
	}

	return nil
}
