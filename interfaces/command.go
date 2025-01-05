package interfaces

type Command interface {
	BuildCommand() (string, error)
	Execute() error
}
