package cmd

import (
	"fmt"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umputun/go-flags"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

func TestOffersUpdatesCommand_Execute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/s/v2/offers/offer", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "{\"results\":["+
			"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
			"	\"main_image\":{\"m_img_375x211\":\"https://example.com/1.jpg\"},"+
			"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
			"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
			"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}},"+
			"{\"id\":2,\"vendor\":{\"slug\":\"property-foo-bar\"},"+
			"	\"main_image\":{\"m_img_375x211\":\"https://example.com/2.jpg\"},"+
			"	\"name\":\"Wille Acme\",\"slug\":\"foo-acme-krakow-zwierzyniec\","+
			"	\"region\":{\"full_name\":\"małopolskie, Kraków, Zwierzyniec\"},"+
			"	\"stats\":{\"ranges_area_max\":373,\"ranges_area_min\":139,\"ranges_price_max\":0,\"ranges_price_min\":0}}],"+
			"\"count\":2,\"page\":1,\"page_size\":2,\"next\":null,\"previous\":null}")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	offerStore, err := store.NewOfferFileStore("mock", &MockEngine{sync.Mutex{}, fstest.MapFS{}})
	require.NoError(t, err)

	notifier := MockNotifier{}

	clock := MockClock{}
	cmd := OffersUpdatesCommand{}
	cmd.SetCommon(CommonOpts{
		PrimaryMarketAPI: api.NewHttpApi(server.URL),
		PrimaryMarketURL: server.URL,
		OfferStore:       offerStore,
		OfferNotifier:    &notifier,
		Clock:            clock,
	})
	p := flags.NewParser(&cmd, flags.Default)
	_, err = p.ParseArgs([]string{
		"--request.regions=1",
	})
	require.NoError(t, err)

	err = cmd.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, notifier.called, []store.Offer{
		{
			Id:            1,
			Slug:          "wille-acme-krakow-bronowice",
			Name:          "Wille Acme",
			VendorSlug:    "bar-sp-z-oo",
			Link:          server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1",
			MainImageLink: "https://example.com/1.jpg",
			ImportedAt:    clock.Now(),
			RegionName:    "małopolskie, Kraków, Bronowice",
			PriceMin:      1450000,
			PriceMax:      1450000,
			AreaMin:       180,
			AreaMax:       180,
		},
		{
			Id:            2,
			Slug:          "foo-acme-krakow-zwierzyniec",
			Name:          "Wille Acme",
			VendorSlug:    "property-foo-bar",
			Link:          server.URL + "/oferty/property-foo-bar/foo-acme-krakow-zwierzyniec-2",
			MainImageLink: "https://example.com/2.jpg",
			ImportedAt:    clock.Now(),
			RegionName:    "małopolskie, Kraków, Zwierzyniec",
			PriceMin:      0,
			PriceMax:      0,
			AreaMin:       139,
			AreaMax:       373,
		},
	})
}

type MockEngine struct {
	m  sync.Mutex
	fs fstest.MapFS
}

func (m *MockEngine) CreateDirectories(_ string) error {
	return nil
}

func (m *MockEngine) PathExists(path string) bool {
	m.m.Lock()
	defer m.m.Unlock()

	return m.fs[path] != nil
}

func (m *MockEngine) Write(path string, bytes []byte) error {
	m.m.Lock()
	defer m.m.Unlock()

	m.fs[path] = &fstest.MapFile{
		Data: bytes,
	}
	return nil
}

func (m *MockEngine) Delete(path string) error {
	m.m.Lock()
	defer m.m.Unlock()

	m.fs[path] = nil
	return nil
}

type MockNotifier struct {
	m      sync.Mutex
	called []store.Offer
}

func (m *MockNotifier) Notify(offer store.Offer) error {
	m.m.Lock()
	defer m.m.Unlock()

	m.called = append(m.called, offer)
	return nil
}

type MockClock struct {
	time time.Time
}

func (m MockClock) Now() time.Time {
	return m.time
}
