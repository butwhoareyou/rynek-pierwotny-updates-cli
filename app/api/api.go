package api

import (
	"encoding/json"
	"errors"
	log "github.com/go-pkgz/lgr"
	"net/http"
	"net/url"
	"strconv"
)

type Api interface {
	GetOffers(request PageableOffersRequest) (*PageableOffers, error)

	GetOffersNextPage(previousPage PageableOffers) (*PageableOffers, error)
}

const (
	PageableOffersRequestSortCreatedDate = "create_date"
)

type PageableOffersRequest struct {
	PageSize int
	Region   int64
	Sort     string
}

type PageableOffers struct {
	Page     int32   `json:"page"`
	PageSize int32   `json:"page_size"`
	Results  []Offer `json:"results"`
	Next     string  `json:"next"`
}

type Offer struct {
	Id        int64          `json:"id"`
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	Vendor    OfferVendor    `json:"vendor"`
	MainImage OfferMainImage `json:"main_image"`
	Region    OfferRegion    `json:"region"`
	Stats     OfferStats     `json:"stats"`
}

type OfferVendor struct {
	Slug string `json:"slug"`
}

type OfferMainImage struct {
	Image375x211 string `json:"m_img_375x211"`
}

type OfferRegion struct {
	FullName string `json:"full_name"`
}

type OfferStats struct {
	RangesAreaMin  int   `json:"ranges_area_min"`
	RangesAreaMax  int   `json:"ranges_area_max"`
	RangesPriceMin int64 `json:"ranges_price_min"`
	RangesPriceMax int64 `json:"ranges_price_max"`
}

const (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
)

type httpApi struct {
	baseUrl    string
	httpClient http.Client
}

func NewHttpApi(baseUrl string) Api {
	return &httpApi{baseUrl: baseUrl, httpClient: http.Client{}}
}

func (api *httpApi) GetOffers(request PageableOffersRequest) (*PageableOffers, error) {
	queryParams := url.Values{}
	queryParams.Add("s", "offer-list")
	queryParams.Add("display_type", "1")
	queryParams.Add("distance", "0")
	queryParams.Add("for_sale", "True")
	queryParams.Add("limited_presentation", "True")
	queryParams.Add("page", "1")
	queryParams.Add("page_size", strconv.Itoa(request.PageSize))
	queryParams.Add("region", strconv.FormatInt(request.Region, 10))
	queryParams.Add("show_on_listing", "True")
	queryParams.Add("sort", request.Sort)
	queryParams.Add("type", "2")

	resp, err := get(&api.httpClient, api.baseUrl+"/s/v2/offers/offer", &queryParams)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("invalid response code from API")
	}

	var pageableOffers PageableOffers
	err = json.NewDecoder(resp.Body).Decode(&pageableOffers)
	return &pageableOffers, err
}

func (api *httpApi) GetOffersNextPage(previousPage PageableOffers) (*PageableOffers, error) {
	resp, err := get(&api.httpClient, previousPage.Next, nil)
	if err != nil {
		return nil, err
	}

	var pageableOffers PageableOffers
	err = json.NewDecoder(resp.Body).Decode(&pageableOffers)
	return &pageableOffers, err
}

func get(client *http.Client, urlStr string, queryParams *url.Values) (*http.Response, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}
	requestUrl := u.String()
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", defaultUserAgent)

	log.Printf("[DEBUG] GET %v ", requestUrl)
	return client.Do(req)
}
