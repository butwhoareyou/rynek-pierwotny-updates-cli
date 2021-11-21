package file

import (
	"fmt"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

const filePermission = 0644

type Engine struct {
	baseDir string
	lock    sync.Mutex
}

func NewSystemEngine(basePath string) (engine.Engine, error) {
	baseDir := strings.TrimRight(basePath, "/")
	eng := Engine{baseDir: baseDir, lock: sync.Mutex{}}
	err := eng.createDirectories(baseDir)
	return &eng, err
}

type NoPathError string

func (e NoPathError) Error() string {
	return fmt.Sprintf("Path %s does not exist", string(e))
}

func (e *Engine) Read(path string) ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	exists, err := e.Exists(path)
	if err != nil {
		return make([]byte, 0), err
	}
	if !exists {
		return make([]byte, 0), NoPathError(path)
	}

	return ioutil.ReadFile(e.concatPath(path))
}

func (e *Engine) Write(path string, bytes []byte) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	err := ioutil.WriteFile(e.concatPath(path), bytes, filePermission)
	return err
}

func (e *Engine) Exists(path string) (bool, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.exists(e.concatPath(path))
}

func (e *Engine) createDirectories(path string) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, err := os.Stat(e.concatPath(path)); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}
	return nil
}

func (e *Engine) concatPath(path string) string {
	return e.baseDir + "/" + path
}

func (e *Engine) exists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}
