package runtime

type Component interface {
	Metadata() Metadata

	Configure(Context) error
	Initialize(Context) error

	Start(Context) error
	Stop(Context) error
}