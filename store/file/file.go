package file

import (
	"io/ioutil"
	"os"
	"sync"
)

type Engine interface {
	CreateDirectories(path string) error
	PathExists(path string) bool
	Write(path string, bytes []byte) error
	Delete(path string) error
}

const filePermission = 0644

type SystemEngine struct {
	lock sync.Mutex
}

func NewSystemEngine() SystemEngine {
	return SystemEngine{lock: sync.Mutex{}}
}

func (e *SystemEngine) CreateDirectories(path string) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}
	return nil
}

func (e *SystemEngine) PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
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
