package mutation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

type AttemptResult struct {
	State            string
	ReasonCode       string
	FailureClass     string
	Terminal         string
	ModelTurns       int
	ToolCalls        int
	MutationAttempts int
	Approval         string
	Lifecycle        []string
	InitialSHA256    string
	ExpectedSHA256   string
	FinalSHA256      string
	WorkspaceDigest  string
	ContextFreshness string
	TemporaryCleanup string
}

type AttemptFunc func(context.Context, int) (AttemptResult, error)

type RunnerOptions struct {
	Profile  Profile
	Gate     string
	Hardware pkgBenchmark.HardwareProfile
	Build    ReportBuild
	Now      func() time.Time
	RunID    func() string
}

func RunGate(ctx context.Context, options RunnerOptions, attempt AttemptFunc) (Report, error) {
	if ctx == nil || attempt == nil {
		return Report{}, errors.New("mutation qualification runner dependencies are nil")
	}
	if err := options.Profile.Validate(); err != nil {
		return Report{}, err
	}
	gate, exists := options.Profile.Gate(options.Gate)
	if !exists {
		return Report{}, fmt.Errorf("mutation qualification gate %q is unknown", options.Gate)
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.RunID == nil {
		options.RunID = randomReportID
	}
	started := options.Now()
	report := Report{
		SchemaVersion: ReportSchemaVersion, RunID: options.RunID(), CreatedAt: started,
		Profile: ReportProfile{
			Version: options.Profile.Version, Digest: options.Profile.Digest(),
			Target: options.Profile.Target.Platform + "/" + options.Profile.Target.Provider,
			Model:  options.Profile.Target.Model,
		},
		Build: options.Build, Hardware: options.Hardware, Gate: gate.ID,
		Scenario: gate.Scenario, Required: gate.RequiredPasses, FailFast: gate.FailFast,
		State: "passed", Samples: make([]Sample, 0, gate.RequiredPasses),
	}
	for index := 1; index <= gate.RequiredPasses; index++ {
		attemptStarted := options.Now()
		attemptContext, cancel := context.WithTimeout(ctx, options.Profile.Protocol.RunDeadline.Duration)
		result, err := attempt(attemptContext, index)
		contextError := attemptContext.Err()
		cancel()
		completed := options.Now()
		if completed.Before(attemptStarted) {
			completed = attemptStarted
		}
		if err != nil || contextError != nil {
			result.State = "failed"
			result.Terminal = "execution_failed"
			result.FailureClass = "harness"
			result.ReasonCode = "attempt_error"
			if contextError != nil {
				result.Terminal = "deadline_exceeded"
				result.ReasonCode = "attempt_deadline"
			}
		}
		sample := sampleFromResult(index, attemptStarted, completed.Sub(attemptStarted), result)
		report.Samples = append(report.Samples, sample)
		if result.State != "passed" {
			report.State = result.State
			break
		}
	}
	completed := options.Now()
	if completed.Before(started) {
		completed = started
	}
	report.CompletedAt = completed
	report.DurationMS = float64(completed.Sub(started)) / float64(time.Millisecond)
	if report.State == "passed" && len(report.Samples) != gate.RequiredPasses {
		report.State = "failed"
	}
	if err := report.Validate(); err != nil {
		return Report{}, err
	}
	return report, nil
}

func sampleFromResult(index int, started time.Time, duration time.Duration, result AttemptResult) Sample {
	return Sample{
		Attempt: index, State: result.State, ReasonCode: result.ReasonCode,
		FailureClass: result.FailureClass, StartedAt: started,
		DurationMS: float64(duration) / float64(time.Millisecond), Terminal: result.Terminal,
		ModelTurns: result.ModelTurns, ToolCalls: result.ToolCalls,
		MutationAttempts: result.MutationAttempts, Approval: result.Approval,
		Lifecycle: append([]string(nil), result.Lifecycle...), InitialSHA256: result.InitialSHA256,
		ExpectedSHA256: result.ExpectedSHA256, FinalSHA256: result.FinalSHA256,
		WorkspaceDigest: result.WorkspaceDigest, ContextFreshness: result.ContextFreshness,
		TemporaryCleanup: result.TemporaryCleanup,
	}
}

func randomReportID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("mutation-%d", time.Now().UnixNano())
	}
	return "mutation-" + hex.EncodeToString(value[:])
}
