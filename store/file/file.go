package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Engine interface {
	CreateDirectories(path string) error
	Read(path string) ([]byte, error)
	Write(path string, bytes []byte) error
	Delete(path string) error
	Exists(path string) bool
}

const filePermission = 0644

type SystemEngine struct {
	lock sync.Mutex
}

func NewSystemEngine() SystemEngine {
	return SystemEngine{lock: sync.Mutex{}}
}

type NoPathError string

func (e NoPathError) Error() string {
	return fmt.Sprintf("Path %s does not exist", string(e))
}

func (e *SystemEngine) CreateDirectories(path string) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}
	return nil
}

func (e *SystemEngine) Read(path string) ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if !e.Exists(path) {
		return make([]byte, 0), NoPathError(path)
	}

	return ioutil.ReadFile(path)
}

func (e *SystemEngine) Write(path string, bytes []byte) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	err := ioutil.WriteFile(path, bytes, filePermission)
	return err
}

func (e *SystemEngine) Delete(path string) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	return os.Remove(path)
}

func (e *SystemEngine) Exists(path string) bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.exists(path)
}

func (e *SystemEngine) exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
