package productconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const maxConfigBytes = 1 << 20

func Load(path string) (Config, error) {
	return load(path, func(config Config) error { return config.Validate() })
}

// LoadChat decodes the same strict versioned document as Load, but validates
// only the common provider/workspace fields and the direct-chat profile. This
// keeps a tool-free chat command independent from optional agent settings.
func LoadChat(path string) (Config, error) {
	return load(path, func(config Config) error {
		if config.Version == Version {
			return config.Validate()
		}
		return config.ValidateChatExecutionProfile()
	})
}

func load(path string, validate func(Config) error) (Config, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read configuration: %w: %w", err, ErrNotFound)
		}
		return Config{}, fmt.Errorf("read configuration: %w: %w", err, ErrInvalid)
	}
	if len(encoded) == 0 || len(encoded) > maxConfigBytes {
		return Config{}, fmt.Errorf("configuration must contain 1..%d bytes: %w", maxConfigBytes, ErrInvalid)
	}
	if err := rejectAliases(encoded); err != nil {
		return Config{}, err
	}
	var header struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(encoded, &header); err != nil {
		return Config{}, fmt.Errorf("decode configuration version: %v: %w", err, ErrInvalid)
	}
	var config Config
	switch header.Version {
	case Version:
		var decoded configV1
		if err := decodeStrict(encoded, &decoded); err != nil {
			return Config{}, err
		}
		config = decoded.config()
	case CandidateVersion:
		var decoded configV2
		if err := decodeStrict(encoded, &decoded); err != nil {
			return Config{}, err
		}
		config = decoded.config()
	default:
		return Config{}, fieldError("version", fmt.Sprintf("must equal %d or %d", Version, CandidateVersion))
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("resolve configuration path: %w: %w", err, ErrInvalid)
	}
	config.path = filepath.Clean(absolute)
	if !filepath.IsAbs(config.Workspace.Root) && config.Workspace.Root != "" {
		config.Workspace.Root = filepath.Clean(filepath.Join(filepath.Dir(config.path), config.Workspace.Root))
	}
	if validate == nil {
		return Config{}, fmt.Errorf("configuration validator is missing: %w", ErrInvalid)
	}
	if err := validate(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

type configV1 struct {
	Version   int             `yaml:"version"`
	Provider  ProviderConfig  `yaml:"provider"`
	Models    ModelsConfig    `yaml:"models"`
	Workspace WorkspaceConfig `yaml:"workspace"`
	Agent     AgentConfig     `yaml:"agent"`
	Policy    PolicyConfig    `yaml:"policy"`
	Limits    LimitsConfig    `yaml:"limits"`
	Context   ContextConfig   `yaml:"context"`
}

func (decoded configV1) config() Config {
	return Config{
		Version: decoded.Version, Provider: decoded.Provider,
		Models: decoded.Models, Workspace: decoded.Workspace,
		Agent: decoded.Agent, Policy: decoded.Policy,
		Limits: decoded.Limits, Context: decoded.Context,
	}
}

type modelsV2 struct {
	Embedding string `yaml:"embedding"`
}

type agentV2 struct {
	ID    string   `yaml:"id"`
	Tools []string `yaml:"tools"`
}

type configV2 struct {
	Version     int               `yaml:"version"`
	Provider    ProviderConfig    `yaml:"provider"`
	Models      modelsV2          `yaml:"models"`
	Workspace   WorkspaceConfig   `yaml:"workspace"`
	Interaction InteractionConfig `yaml:"interaction"`
	Agent       agentV2           `yaml:"agent"`
	Policy      PolicyConfig      `yaml:"policy"`
	Limits      LimitsConfig      `yaml:"limits"`
	Context     ContextConfig     `yaml:"context"`
}

func (decoded configV2) config() Config {
	return Config{
		Version: decoded.Version, Provider: decoded.Provider,
		Models: ModelsConfig{
			Chat: decoded.Interaction.Agent.Model, Embedding: decoded.Models.Embedding,
		},
		Workspace: decoded.Workspace, Interaction: decoded.Interaction,
		Agent: AgentConfig{
			ID: decoded.Agent.ID, Streaming: decoded.Interaction.Agent.Streaming,
			Tools: decoded.Agent.Tools,
		},
		Policy: decoded.Policy, Limits: decoded.Limits, Context: decoded.Context,
	}
}

func decodeStrict(encoded []byte, destination any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode configuration: %v: %w", err, ErrInvalid)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("configuration contains multiple YAML documents: %w", ErrInvalid)
		}
		return fmt.Errorf("decode trailing configuration: %v: %w", err, ErrInvalid)
	}
	return nil
}

func rejectAliases(encoded []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	for {
		var document yaml.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse configuration: %v: %w", err, ErrInvalid)
		}
		if containsAlias(&document) {
			return fmt.Errorf("configuration aliases are not supported: %w", ErrInvalid)
		}
	}
}

func containsAlias(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return true
	}
	for _, child := range node.Content {
		if containsAlias(child) {
			return true
		}
	}
	return false
}
