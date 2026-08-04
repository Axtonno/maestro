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

type Transition struct {
	From State
	To State
}

var ValidTransitions = []Transition{
	{StateCreated, StateConfigured},
	{StateConfigured, StateInitialized},
	{StateInitialized, StateRunning},
	{StateRunning, StateStopping},
	{StateStopping, StateStopped},

	{StateConfigured, StateFailed},
	{StateInitialized, StateFailed},
	{StateRunning, StateFailed},
}