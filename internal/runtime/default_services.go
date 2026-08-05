package runtime

import "reflect"

type emptyConfig struct{}

func newEmptyConfig() *emptyConfig {
	return &emptyConfig{}
}

func (c *emptyConfig) Get(_ string) any {
	return nil
}

type noopLogger struct{}

func newNoopLogger() *noopLogger {
	return &noopLogger{}
}

func (l *noopLogger) Debug(_ string) {}

func (l *noopLogger) Info(_ string) {}

func (l *noopLogger) Warn(_ string) {}

func (l *noopLogger) Error(_ string) {}

func nilService(service any) bool {
	if service == nil {
		return true
	}

	value := reflect.ValueOf(service)

	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
