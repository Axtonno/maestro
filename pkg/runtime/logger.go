package runtime

type Logger interface {
	Debug(msg string)

	Info(msg string)

	Warn(msg string)

	Error(msg string)
}