package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/antonio-cafeo/maestro/internal/benchmark/mutation"
)

func runBenchMutation(arguments []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("maestro bench mutation", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profilePath := flags.String(
		"profile",
		"docs/mutation-qualification-profile.yaml",
		"path to the mutation qualification profile",
	)
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro bench mutation does not accept positional arguments")
		return 2
	}
	profile, err := mutation.LoadProfile(*profilePath)
	if err != nil {
		fmt.Fprintf(stderr, "load mutation qualification profile: %v\n", err)
		return 1
	}
	fixture, err := mutation.MaterializeFixture(context.Background(), profile)
	if err != nil {
		fmt.Fprintf(stderr, "validate mutation qualification fixture: %v\n", err)
		return 1
	}
	if err := fixture.Cleanup(); err != nil {
		fmt.Fprintln(stderr, "validate mutation qualification fixture: cleanup_failed")
		return 1
	}
	fmt.Fprintf(stdout, "mutation qualification profile valid: version=%d gates=%d scenarios=%d\n",
		profile.Version, len(profile.Gates), len(profile.MutationMatrix))
	fmt.Fprintf(stdout, "profile_sha256\t%s\n", profile.Digest())
	fmt.Fprintf(stdout, "candidate\t%s/%s\t%s\n", profile.Target.Platform, profile.Target.Provider, profile.Target.Model)
	return 0
}
