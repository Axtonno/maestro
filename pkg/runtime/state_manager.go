package runtime


type StateManager interface {

	Get(Component) ComponentState

	Set(Component, State)

	Fail(Component, error)

}