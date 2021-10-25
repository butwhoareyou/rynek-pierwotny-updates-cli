package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpApi_GetOffers(t *testing.T) {
	mockHttpResponseBody := "{\"results\":[" +
		"{\"id\":1,\"vendor\":{\"slug\":\"property-foo-bar\"},\"main_image\":{\"m_img_375x211\":\"https://example.com/1.jpg\"},\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-zwierzyniec\"}," +
		"{\"id\":2,\"vendor\":{\"slug\":\"foo-bar-developer\"},\"main_image\":{\"m_img_375x211\":\"https://example.com/2.jpg\"},\"name\":\"Baz House\",\"slug\":\"baz-house-foo-kokotow\"}]}"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, mockHttpResponseBody)
	}))
	defer server.Close()
	api := NewHttpApi(server.URL)
	request := PageableOffersRequest{Region: 52258, Sort: PageableOffersRequestSortCreatedDate}

	resp, err := api.GetOffers(request)

	require.NoError(t, err)
	expected := &PageableOffers{Results: []Offer{
		{
			Id:        1,
			Name:      "Wille Acme",
			Slug:      "wille-acme-krakow-zwierzyniec",
			Vendor:    OfferVendor{Slug: "property-foo-bar"},
			MainImage: OfferMainImage{Image375x211: "https://example.com/1.jpg"},
		},
		{
			Id:        2,
			Name:      "Baz House",
			Slug:      "baz-house-foo-kokotow",
			Vendor:    OfferVendor{Slug: "foo-bar-developer"},
			MainImage: OfferMainImage{Image375x211: "https://example.com/2.jpg"},
		},
	}}

	assert.Equal(t, resp, expected)
}
