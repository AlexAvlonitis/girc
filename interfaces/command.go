package interfaces

type Command interface {
	Print() (string, error)
	Execute() error
}
