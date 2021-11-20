package cmd

import (
	"fmt"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/file"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer"
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
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/s/v2/offers/offer", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprintf(w, "{\"results\":["+
			"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
			"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
			"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
			"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
			"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}},"+
			"{\"id\":2,\"vendor\":{\"slug\":\"property-foo-bar\"},"+
			"	\"main_image\":{\"m_img_375x211\":\"%s/2.jpg\"},"+
			"	\"name\":\"Wille Acme\",\"slug\":\"foo-acme-krakow-zwierzyniec\","+
			"	\"region\":{\"full_name\":\"małopolskie, Kraków, Zwierzyniec\"},"+
			"	\"stats\":{\"ranges_area_max\":373,\"ranges_area_min\":139,\"ranges_price_max\":0,\"ranges_price_min\":0}}],"+
			"\"count\":2,\"page\":1,\"page_size\":2,\"next\":null,\"previous\":null}",
			server.URL, server.URL)
	})
	mux.HandleFunc("/1.jpg", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "yay")
	})
	mux.HandleFunc("/2.jpg", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "yey")
	})

	offerStore, err := store.NewOfferFileStore("mock", &MockEngine{sync.Mutex{}, fstest.MapFS{}})
	require.NoError(t, err)

	notifier := MockWriter{}
	httpClient := http.Client{}
	clock := MockClock{}
	cmd := OffersUpdatesCommand{}
	cmd.SetCommon(CommonOpts{
		PrimaryMarketAPI: api.NewHttpApi(server.URL),
		PrimaryMarketURL: server.URL,
		OfferStore:       offerStore,
		OfferWriter:      &notifier,
		Clock:            clock,
		HttpClient:       httpClient,
	})
	p := flags.NewParser(&cmd, flags.Default)
	_, err = p.ParseArgs([]string{
		"--request.regions=1",
	})
	require.NoError(t, err)

	err = cmd.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, notifier.called, []writer.Message{
		{
			Title: server.URL + "/1.jpg",
			Image: []byte("yay"),
			Text:  "🏡Wille Acme\n📍 małopolskie, Kraków, Bronowice\n📏 180-180\n🙀 1450000-1450000\n\n➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1",
		},
		{
			Title: server.URL + "/2.jpg",
			Image: []byte("yey"),
			Text:  "🏡Wille Acme\n📍 małopolskie, Kraków, Zwierzyniec\n📏 139-373\n🙀 0-0\n\n➡️ " + server.URL + "/oferty/property-foo-bar/foo-acme-krakow-zwierzyniec-2",
		},
	})
}

func TestOffersUpdatesCommand_Execute_NoPriceChange(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	requestIdx := 0
	mux.HandleFunc("/s/v2/offers/offer", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		if requestIdx == 0 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}
		if requestIdx == 1 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}

		requestIdx++
	})
	mux.HandleFunc("/1.jpg", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "yay")
	})

	offerStore, err := store.NewOfferFileStore("mock", &MockEngine{sync.Mutex{}, fstest.MapFS{}})
	require.NoError(t, err)

	notifier := MockWriter{}
	httpClient := http.Client{}
	clock := MockClock{}
	cmd := OffersUpdatesCommand{}
	cmd.SetCommon(CommonOpts{
		PrimaryMarketAPI: api.NewHttpApi(server.URL),
		PrimaryMarketURL: server.URL,
		OfferStore:       offerStore,
		OfferWriter:      &notifier,
		Clock:            clock,
		HttpClient:       httpClient,
	})
	p := flags.NewParser(&cmd, flags.Default)
	_, err = p.ParseArgs([]string{
		"--request.regions=1",
	})
	require.NoError(t, err)

	err = cmd.Execute(nil)
	require.NoError(t, err)

	// Request again and receive changed price
	err = cmd.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, notifier.called, []writer.Message{
		{
			Title: server.URL + "/1.jpg",
			Image: []byte("yay"),
			Text:  "🏡Wille Acme\n📍 małopolskie, Kraków, Bronowice\n📏 180-180\n🙀 1450000-1450000\n\n➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1",
		},
	})
}

func TestOffersUpdatesCommand_Execute_PriceChange_Rise(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	requestIdx := 0
	mux.HandleFunc("/s/v2/offers/offer", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		if requestIdx == 0 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}
		if requestIdx == 1 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1750000,\"ranges_price_min\":1550000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}

		requestIdx++
	})
	mux.HandleFunc("/1.jpg", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "yay")
	})

	offerStore, err := store.NewOfferFileStore("mock", &MockEngine{sync.Mutex{}, fstest.MapFS{}})
	require.NoError(t, err)

	notifier := MockWriter{}
	httpClient := http.Client{}
	clock := MockClock{}
	cmd := OffersUpdatesCommand{}
	cmd.SetCommon(CommonOpts{
		PrimaryMarketAPI: api.NewHttpApi(server.URL),
		PrimaryMarketURL: server.URL,
		OfferStore:       offerStore,
		OfferWriter:      &notifier,
		Clock:            clock,
		HttpClient:       httpClient,
	})
	p := flags.NewParser(&cmd, flags.Default)
	_, err = p.ParseArgs([]string{
		"--request.regions=1",
	})
	require.NoError(t, err)

	err = cmd.Execute(nil)
	require.NoError(t, err)

	// Request again and receive changed price
	err = cmd.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, notifier.called, []writer.Message{
		{
			Title: server.URL + "/1.jpg",
			Image: []byte("yay"),
			Text:  "🏡Wille Acme\n📍 małopolskie, Kraków, Bronowice\n📏 180-180\n🙀 1450000-1450000\n\n➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1",
		},
		{
			Title: "",
			Image: make([]byte, 0),
			Text:  "➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1\n\n↗️ 1550000-1750000",
		},
	})
}

func TestOffersUpdatesCommand_Execute_PriceChange_Drop(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	requestIdx := 0
	mux.HandleFunc("/s/v2/offers/offer", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		if requestIdx == 0 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1450000,\"ranges_price_min\":1450000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}
		if requestIdx == 1 {
			_, _ = fmt.Fprintf(w, "{\"results\":["+
				"{\"id\":1,\"vendor\":{\"slug\":\"bar-sp-z-oo\"},"+
				"	\"main_image\":{\"m_img_375x211\":\"%s/1.jpg\"},"+
				"	\"name\":\"Wille Acme\",\"slug\":\"wille-acme-krakow-bronowice\","+
				"	\"region\":{\"full_name\":\"małopolskie, Kraków, Bronowice\"},"+
				"	\"stats\":{\"ranges_area_max\":180,\"ranges_area_min\":180,\"ranges_price_max\":1150000,\"ranges_price_min\":950000}}],"+
				"\"count\":1,\"page\":1,\"page_size\":1,\"next\":null,\"previous\":null}",
				server.URL)
		}

		requestIdx++
	})
	mux.HandleFunc("/1.jpg", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "yay")
	})

	offerStore, err := store.NewOfferFileStore("mock", &MockEngine{sync.Mutex{}, fstest.MapFS{}})
	require.NoError(t, err)

	notifier := MockWriter{}
	httpClient := http.Client{}
	clock := MockClock{}
	cmd := OffersUpdatesCommand{}
	cmd.SetCommon(CommonOpts{
		PrimaryMarketAPI: api.NewHttpApi(server.URL),
		PrimaryMarketURL: server.URL,
		OfferStore:       offerStore,
		OfferWriter:      &notifier,
		Clock:            clock,
		HttpClient:       httpClient,
	})
	p := flags.NewParser(&cmd, flags.Default)
	_, err = p.ParseArgs([]string{
		"--request.regions=1",
	})
	require.NoError(t, err)

	err = cmd.Execute(nil)
	require.NoError(t, err)

	// Request again and receive changed price
	err = cmd.Execute(nil)
	require.NoError(t, err)

	assert.Equal(t, notifier.called, []writer.Message{
		{
			Title: server.URL + "/1.jpg",
			Image: []byte("yay"),
			Text:  "🏡Wille Acme\n📍 małopolskie, Kraków, Bronowice\n📏 180-180\n🙀 1450000-1450000\n\n➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1",
		},
		{
			Title: "",
			Image: make([]byte, 0),
			Text:  "➡️ " + server.URL + "/oferty/bar-sp-z-oo/wille-acme-krakow-bronowice-1\n\n↘️ 950000-1150000",
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

func (m *MockEngine) Read(path string) ([]byte, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.fs[path] == nil {
		return make([]byte, 0), file.NoPathError(path)
	}

	return m.fs[path].Data, nil
}

func (m *MockEngine) Exists(path string) bool {
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

type MockWriter struct {
	m      sync.Mutex
	called []writer.Message
}

func (m *MockWriter) Write(offer writer.Message) error {
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
