package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerIsFailFastAndReportRoundTrips(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	report, err := RunGate(context.Background(), RunnerOptions{
		Profile: profile, Gate: GateA, Now: func() time.Time { return now },
		RunID: func() string { return "mutation-test" },
	}, func(_ context.Context, attempt int) (AttemptResult, error) {
		result := validAttempt(profile)
		if attempt == 2 {
			result.State = "failed"
			result.Terminal = "provider_failure"
			result.ReasonCode = "tool_call_count"
			result.FailureClass = "model"
		}
		return result, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.State != "failed" || len(report.Samples) != 2 {
		t.Fatalf("runner did not stop fail-fast: %#v", report)
	}
	encoded, err := JSONBytes(report)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeReportJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.RunID != report.RunID || len(decoded.Samples) != 2 {
		t.Fatalf("report round trip mismatch: %#v", decoded)
	}
	var markdown bytes.Buffer
	if err := EncodeReportMarkdown(&markdown, decoded); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"return response", "/home/", "prompt"} {
		if strings.Contains(markdown.String(), forbidden) {
			t.Fatalf("Markdown leaked %q", forbidden)
		}
	}
}

func TestReportWriterUsesPrivatePermissions(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	report, err := RunGate(context.Background(), RunnerOptions{
		Profile: profile, Gate: GateB, Now: func() time.Time { return now },
		RunID: func() string { return "mutation-private" },
	}, func(context.Context, int) (AttemptResult, error) { return validAttempt(profile), nil })
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "report.json")
	if err := WriteReport(name, func(writer io.Writer) error { return EncodeReportJSON(writer, report) }); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected report permissions: %o", info.Mode().Perm())
	}
}

func TestPublishedMutationReportSchemaIsValidJSON(t *testing.T) {
	content, err := os.ReadFile("../../../docs/schemas/mutation-qualification-report-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatal(err)
	}
	if schema["title"] != "Maestro Mutation Qualification Report v1" {
		t.Fatalf("unexpected schema identity: %#v", schema["title"])
	}
}

func validAttempt(profile Profile) AttemptResult {
	digest := strings.Repeat("a", 64)
	return AttemptResult{
		State: "passed", Terminal: "completed", Approval: "not_requested",
		InitialSHA256:  profile.Fixture.InitialSHA256,
		ExpectedSHA256: profile.Fixture.ExpectedSHA256,
		FinalSHA256:    profile.Fixture.InitialSHA256, WorkspaceDigest: digest,
		ContextFreshness: "not_applicable", TemporaryCleanup: "clean",
	}
}
