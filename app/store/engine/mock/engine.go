package mock

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/file"
)

type Engine struct {
}

func NewEngine() engine.Engine {
	return &Engine{}
}

func (e Engine) Read(path string) ([]byte, error) {
	return make([]byte, 0), file.NoPathError(path)
}

func (e Engine) Write(_ string, _ []byte) error {
	return nil
}

func (e Engine) Exists(_ string) (bool, error) {
	return false, nil
}
