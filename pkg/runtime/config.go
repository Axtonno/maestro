package runtime

type Config interface {
	Get(string) any
}

type valueConfig struct {
	values map[string]any
}

// NewConfig returns a read-only configuration snapshot with exact,
// case-sensitive keys. The input map is copied during construction.
func NewConfig(values map[string]any) Config {
	copiedValues := make(map[string]any, len(values))

	for key, value := range values {
		copiedValues[key] = value
	}

	return &valueConfig{
		values: copiedValues,
	}
}

func (c *valueConfig) Get(key string) any {
	return c.values[key]
}
