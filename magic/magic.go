package magic

type Magic interface {
	Name() string
	Trigger() error
}
