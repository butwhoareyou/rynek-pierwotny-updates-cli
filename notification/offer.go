package notification

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	log "github.com/go-pkgz/lgr"
)

type OfferNotifier interface {
	Notify(offer store.Offer) error
}

type LogNotifier struct{}

func (l *LogNotifier) Notify(offer store.Offer) error {
	log.Printf("[INFO] Notifying offer %v", offer)
	return nil
}
