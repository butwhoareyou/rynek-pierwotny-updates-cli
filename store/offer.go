package store

import (
	"encoding/json"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/file"
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

type OfferStore interface {
	Create(offer Offer) error

	Exists(offerId int64) (bool, error)

	Delete(offer Offer) error
}

func NewOfferFileStore(baseDir string, engine file.Engine) (OfferStore, error) {
	fileStore := OfferFileStore{baseDir: baseDir, engine: engine}
	err := fileStore.initialize()
	return &fileStore, err
}

type OfferFileStore struct {
	baseDir string
	engine  file.Engine
}

func (f *OfferFileStore) Exists(offerId int64) (bool, error) {
	exists := f.engine.PathExists(f.fileName(offerId))
	return exists, nil
}

func (f *OfferFileStore) Create(offer Offer) error {
	fileName := f.fileName(offer.Id)
	content, err := f.serialize(offer)
	if err != nil {
		return err
	}
	err = f.engine.Write(fileName, content)
	return err
}

func (f *OfferFileStore) Delete(offer Offer) error {
	fileName := f.fileName(offer.Id)
	err := f.engine.Delete(fileName)
	return err
}

func (f *OfferFileStore) initialize() error {
	return f.engine.CreateDirectories(f.baseDir)
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
