package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
)

const profileIdentitySchemaVersion = 1

type profileIdentity struct {
	SchemaVersion int    `json:"schema_version"`
	Profile       string `json:"profile"`
	Provider      string `json:"provider"`
	ChatModel     string `json:"chat_model"`
	MutationModel string `json:"mutation_model"`
}

func runProfile(arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	flags := flag.NewFlagSet("maestro profile", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}
	configPath := flags.String("config", "", "path to Maestro configuration")
	workspaceCurrent := flags.Bool("workspace-current", false, "use the current working directory as workspace root")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprintln(stdout, "usage: maestro profile [--config path] [--workspace-current]")
			return 0
		}
		fmt.Fprintln(stderr, "profile failed: invalid_request")
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "profile failed: invalid_request")
		return 2
	}
	config, err := resolveAndLoadMutation(*configPath, dependencies)
	if err == nil {
		config, err = applyCurrentWorkspace(config, *workspaceCurrent, dependencies, productconfig.Config.ValidateMutationExecutionProfile)
	}
	if err != nil {
		fmt.Fprintln(stderr, "profile failed: invalid_request")
		renderConfigurationDiagnostic(stderr, err)
		return 2
	}
	identity := profileIdentity{
		SchemaVersion: profileIdentitySchemaVersion,
		Profile:       string(config.ProductProfile()),
		Provider:      config.Provider.ID,
		ChatModel:     config.DirectChat.Model,
		MutationModel: config.ControlledMutation.Model,
	}
	if err := json.NewEncoder(stdout).Encode(identity); err != nil {
		fmt.Fprintln(stderr, "profile failed: output_unavailable")
		return 1
	}
	return 0
}
