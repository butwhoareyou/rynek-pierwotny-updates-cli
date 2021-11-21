package engine

type Engine interface {
	Read(path string) ([]byte, error)
	Write(path string, bytes []byte) error
	Exists(path string) (bool, error)
}
