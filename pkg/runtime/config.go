package runtime

type Config interface {
	Get(string) any
}
