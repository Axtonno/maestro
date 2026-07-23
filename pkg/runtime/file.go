package runtime

type Config interface {
	Get(key string) any
}