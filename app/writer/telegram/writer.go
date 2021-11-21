package telegram

import (
	"fmt"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer"
	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Writer struct {
	ChatId int64
	BotAPI *tgbotapi.BotAPI
}

func NewWriter(chatId int64, api *tgbotapi.BotAPI) *Writer {
	return &Writer{ChatId: chatId, BotAPI: api}
}

func (w *Writer) Write(offer writer.Message) error {
	if len(offer.Image) > 0 {
		return w.photoUpload(offer)
	}
	if len(offer.Text) > 0 {
		return w.message(offer)
	}
	return fmt.Errorf("no handler for %v", offer)
}

func (w *Writer) photoUpload(offer writer.Message) error {
	image := tgbotapi.FileBytes{
		Name:  offer.Title,
		Bytes: offer.Image,
	}
	upload := tgbotapi.NewPhotoUpload(w.ChatId, image)
	upload.Caption = offer.Text

	log.Printf("[DEBUG] Uploading image %v..", upload)

	_, err := w.BotAPI.Send(upload)
	return err
}

func (w *Writer) message(offer writer.Message) error {
	txt := offer.Text
	if len(offer.Title) > 0 {
		txt = offer.Title + "\n\n" + txt
	}
	msg := tgbotapi.NewMessage(w.ChatId, txt)

	log.Printf("[DEBUG] Sending text message %v..", msg)

	_, err := w.BotAPI.Send(msg)
	return err
}
