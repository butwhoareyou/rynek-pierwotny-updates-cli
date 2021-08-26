package telegram

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"strconv"
)

type Notifier struct {
	ChatId     int64
	BotAPI     *tgbotapi.BotAPI
	httpClient http.Client
}

func NewNotifier(chatId int64, api *tgbotapi.BotAPI, client http.Client) *Notifier {
	return &Notifier{ChatId: chatId, BotAPI: api, httpClient: client}
}

func (n *Notifier) Notify(offer store.Offer) error {
	image, err := n.offerImage(offer)
	if err != nil {
		return err
	}
	upload := tgbotapi.NewPhotoUpload(n.ChatId, image)
	upload.Caption = n.offerMessageText(offer)

	log.Printf("[DEBUG] Uploading image %v..", upload)

	_, err = n.BotAPI.Send(upload)
	return err
}

func (n *Notifier) offerMessageText(offer store.Offer) string {
	return "🏡" + offer.Name +
		"\n📍 " + offer.RegionName +
		"\n📏 " + strconv.Itoa(offer.AreaMin) + "-" + strconv.Itoa(offer.AreaMax) +
		"\n🙀 " + strconv.FormatInt(offer.PriceMin, 10) + "-" + strconv.FormatInt(offer.PriceMax, 10) +
		"\n" +
		"\n➡️ " + offer.Link
}

func (n *Notifier) offerImage(offer store.Offer) (tgbotapi.FileReader, error) {
	resp, err := n.httpClient.Get(offer.MainImageLink)
	if err != nil {
		return tgbotapi.FileReader{}, err
	}

	fileReader := tgbotapi.FileReader{
		Name:   offer.MainImageLink,
		Reader: resp.Body,
		Size:   resp.ContentLength,
	}

	return fileReader, nil

}
