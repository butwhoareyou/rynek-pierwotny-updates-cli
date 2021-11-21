package writer

import (
	log "github.com/go-pkgz/lgr"
)

type Message struct {
	Title string
	Image []byte
	Text  string
}

type MessageWriter interface {
	Write(message Message) error
}

type LogWriter struct{}

func (l *LogWriter) Write(message Message) error {
	log.Printf("[INFO] Notifying message with title %v", message.Title)
	return nil
}
