package productconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

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
			return Config{}, withDiagnostic(fmt.Errorf("read configuration: %w: %w", err, ErrNotFound), DiagnosticReadFailed, "")
		}
		return Config{}, withDiagnostic(fmt.Errorf("read configuration: %w: %w", err, ErrInvalid), DiagnosticReadFailed, "")
	}
	if len(encoded) == 0 || len(encoded) > maxConfigBytes {
		return Config{}, withDiagnostic(fmt.Errorf("configuration must contain 1..%d bytes: %w", maxConfigBytes, ErrInvalid), DiagnosticYAMLInvalid, "")
	}
	if err := rejectAliases(encoded); err != nil {
		return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
	}
	var header struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(encoded, &header); err != nil {
		return Config{}, withDiagnostic(fmt.Errorf("decode configuration version: %v: %w", err, ErrInvalid), DiagnosticYAMLInvalid, "")
	}
	var config Config
	switch header.Version {
	case Version:
		var decoded configV1
		if field, err := unknownYAMLField(encoded, reflect.TypeOf(decoded)); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		} else if field != "" {
			return Config{}, withDiagnostic(fmt.Errorf("configuration contains unknown field: %w", ErrInvalid), DiagnosticUnknownField, field)
		}
		if err := decodeStrict(encoded, &decoded); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		}
		config = decoded.config()
	case CandidateVersion:
		var decoded configV2
		if field, err := unknownYAMLField(encoded, reflect.TypeOf(decoded)); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		} else if field != "" {
			return Config{}, withDiagnostic(fmt.Errorf("configuration contains unknown field: %w", ErrInvalid), DiagnosticUnknownField, field)
		}
		if err := decodeStrict(encoded, &decoded); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		}
		config = decoded.config()
	case QualificationVersion:
		var decoded configV3
		if field, err := unknownYAMLField(encoded, reflect.TypeOf(decoded)); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		} else if field != "" {
			return Config{}, withDiagnostic(fmt.Errorf("configuration contains unknown field: %w", ErrInvalid), DiagnosticUnknownField, field)
		}
		if err := decodeStrict(encoded, &decoded); err != nil {
			return Config{}, withDiagnostic(err, DiagnosticYAMLInvalid, "")
		}
		config = decoded.config()
	default:
		err := fieldError("version", fmt.Sprintf("must equal %d, %d or %d", Version, CandidateVersion, QualificationVersion))
		kind := DiagnosticInvalidValue
		if !hasYAMLPath(encoded, "version") {
			kind = DiagnosticMissingField
		}
		return Config{}, withDiagnostic(err, kind, "version")
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
		field := validationField(err)
		kind := DiagnosticInvalidValue
		if field != "" && !hasYAMLPath(encoded, field) {
			kind = DiagnosticMissingField
		}
		return Config{}, withDiagnostic(err, kind, field)
	}
	return config, nil
}

func unknownYAMLField(encoded []byte, target reflect.Type) (string, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(encoded, &document); err != nil {
		return "", fmt.Errorf("parse configuration fields: %v: %w", err, ErrInvalid)
	}
	if len(document.Content) == 0 {
		return "", nil
	}
	return findUnknownYAMLField(document.Content[0], target, ""), nil
}

func findUnknownYAMLField(node *yaml.Node, target reflect.Type, prefix string) string {
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	if node == nil || node.Kind != yaml.MappingNode || target.Kind() != reflect.Struct {
		return ""
	}
	fields := yamlStructFields(target)
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := node.Content[index].Value
		child := node.Content[index+1]
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		childType, exists := fields[key]
		if !exists {
			return path
		}
		if unknown := findUnknownYAMLField(child, childType, path); unknown != "" {
			return unknown
		}
	}
	return ""
}

func yamlStructFields(target reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type)
	for index := 0; index < target.NumField(); index++ {
		field := target.Field(index)
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get("yaml")
		parts := strings.Split(tag, ",")
		name := parts[0]
		inline := slicesContain(parts[1:], "inline")
		if inline {
			for childName, childType := range yamlStructFields(field.Type) {
				fields[childName] = childType
			}
			continue
		}
		if name == "" || name == "-" {
			continue
		}
		fields[name] = field.Type
	}
	return fields
}

func slicesContain(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasYAMLPath(encoded []byte, field string) bool {
	parts := strings.Split(field, ".")
	var document yaml.Node
	if err := yaml.Unmarshal(encoded, &document); err != nil || len(document.Content) == 0 {
		return false
	}
	node := document.Content[0]
	for _, part := range parts {
		if index := strings.IndexByte(part, '['); index >= 0 {
			part = part[:index]
		}
		if part == "" || node.Kind != yaml.MappingNode {
			return false
		}
		found := false
		for index := 0; index+1 < len(node.Content); index += 2 {
			if node.Content[index].Value == part {
				node = node.Content[index+1]
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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

type chatProfileV2 struct {
	ProfileConfig  `yaml:",inline"`
	MaxFileBytes   int `yaml:"max_file_bytes"`
	MaxOutputBytes int `yaml:"max_output_bytes"`
}

type interactionV2 struct {
	Chat  chatProfileV2      `yaml:"chat"`
	Agent AgentProfileConfig `yaml:"agent"`
}

type configV2 struct {
	Version     int             `yaml:"version"`
	Provider    ProviderConfig  `yaml:"provider"`
	Models      modelsV2        `yaml:"models"`
	Workspace   WorkspaceConfig `yaml:"workspace"`
	Interaction interactionV2   `yaml:"interaction"`
	Agent       agentV2         `yaml:"agent"`
	Policy      PolicyConfig    `yaml:"policy"`
	Limits      LimitsConfig    `yaml:"limits"`
	Context     ContextConfig   `yaml:"context"`
}

func (decoded configV2) config() Config {
	interaction := InteractionConfig{
		Chat: ChatProfileConfig{
			ProfileConfig:  decoded.Interaction.Chat.ProfileConfig,
			MaxFileBytes:   decoded.Interaction.Chat.MaxFileBytes,
			MaxOutputBytes: decoded.Interaction.Chat.MaxOutputBytes,
		},
		Agent: decoded.Interaction.Agent,
	}
	return Config{
		Version: decoded.Version, Provider: decoded.Provider,
		Models: ModelsConfig{
			Chat: interaction.Agent.Model, Embedding: decoded.Models.Embedding,
		},
		Workspace: decoded.Workspace, Interaction: interaction,
		Agent: AgentConfig{
			ID: decoded.Agent.ID, Streaming: interaction.Agent.Streaming,
			Tools: decoded.Agent.Tools,
		},
		Policy: decoded.Policy, Limits: decoded.Limits, Context: decoded.Context,
	}
}

type configV3 struct {
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

func (decoded configV3) config() Config {
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
