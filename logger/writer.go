package logger

type IWriter interface {
	write(msg *message)
	close()
}
