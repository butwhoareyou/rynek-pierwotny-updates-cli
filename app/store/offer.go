package store

import (
	"bytes"
	"encoding/json"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine"
	"strconv"
	"time"
)

type Offer struct {
	Id            int64     `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	VendorSlug    string    `json:"vendor_slug"`
	Link          string    `json:"link"`
	MainImageLink string    `json:"main_image_link"`
	ImportedAt    time.Time `json:"imported_at"`
	RegionName    string    `json:"region_name"`
	PriceMin      int64     `json:"price_min"`
	PriceMax      int64     `json:"price_max"`
	AreaMin       int       `json:"area_min"`
	AreaMax       int       `json:"area_max"`
}

func (t *Offer) CompareAveragePrices(o Offer) int {
	dif := t.PriceMin + t.PriceMax - (o.PriceMin + o.PriceMax)
	if dif > 0 {
		return 1
	}
	if dif < 0 {
		return -1
	}
	return 0
}

type OfferStore interface {
	Save(offer Offer) error

	Get(offerId int64) (Offer, error)
}

func NewOfferFileStore(engine engine.Engine) OfferStore {
	fileStore := OfferFileStore{engine: engine}
	return &fileStore
}

type OfferFileStore struct {
	engine engine.Engine
}

func (f *OfferFileStore) Get(offerId int64) (Offer, error) {
	b, err := f.engine.Read(f.fileName(offerId))
	if err != nil {
		return Offer{}, err
	}
	return f.deserialize(b)
}

func (f *OfferFileStore) Save(offer Offer) error {
	fileName := f.fileName(offer.Id)
	content, err := f.serialize(offer)
	if err != nil {
		return err
	}
	err = f.engine.Write(fileName, content)
	return err
}

func (f *OfferFileStore) fileName(offerId int64) string {
	return strconv.FormatInt(offerId, 10) + ".json"
}

func (f *OfferFileStore) serialize(serializable interface{}) ([]byte, error) {
	content, err := json.Marshal(serializable)
	if err != nil {
		return []byte{}, err
	}
	return content, err
}

func (f *OfferFileStore) deserialize(b []byte) (Offer, error) {
	var offer Offer
	r := bytes.NewReader(b)
	err := json.NewDecoder(r).Decode(&offer)
	return offer, err
}
