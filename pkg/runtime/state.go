package runtime

type State uint8

const (
	StateUnknown State = iota

	StateCreated

	StateConfigured

	StateInitialized

	StateRunning

	StateStopping

	StateStopped

	StateFailed
)