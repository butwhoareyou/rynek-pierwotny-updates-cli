package cmd

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	file "github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/file"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer"
	log "github.com/go-pkgz/lgr"
	"go.uber.org/multierr"
	"io/ioutil"
	"strconv"
	"sync"
)

type OffersUpdatesCommand struct {
	PropertiesRequest struct {
		Regions []int64 `short:"r" long:"regions" env:"REGIONS" description:"offer regions"`
	} `group:"request" namespace:"request" env-namespace:"REQUEST"`
	CommonOpts
}

func (cmd *OffersUpdatesCommand) Execute(_ []string) error {
	resetEnv("TELEGRAM_CHAT_ID", "TELEGRAM_TOKEN", "AWS_ACCESS_KEY", "AWS_SECRET_KET")

	log.Printf("[DEBUG] Executing offers updates command..")

	doneCh := make(chan bool)
	errCh := make(chan error)
	defer func() {
		close(doneCh)
		close(errCh)
	}()

	var err error
	cmd.doExecute(doneCh, errCh)

	for {
		select {
		case nrErr := <-errCh:
			err = multierr.Append(err, nrErr)
		case <-doneCh:
			return err
		}
	}
}

func (cmd *OffersUpdatesCommand) doExecute(doneCh chan<- bool, errCh chan<- error) {
	go func() {

		newOffersCh, priseRiseCh, priseDropCh := cmd.orchestrateOffers(errCh,
			cmd.mapApiOffers(errCh,
				cmd.fetchOffers(errCh,
					cmd.streamRegions())))

		persistOffersCh := merge(
			cmd.writeNewOffers(errCh, newOffersCh),
			cmd.writeOffersPriceRise(errCh, priseRiseCh),
			cmd.writeOffersPriceDrop(errCh, priseDropCh),
		)

		cmd.persistOffers(doneCh, errCh, persistOffersCh)
	}()
}

// streamRegions sends all requested regions to the channel
func (cmd *OffersUpdatesCommand) streamRegions() <-chan int64 {
	regionsCh := make(chan int64)
	go func() {
		defer close(regionsCh)

		for _, r := range cmd.PropertiesRequest.Regions {
			regionsCh <- r
		}
	}()
	return regionsCh
}

// fetchOffers performs api call to get all available offers for specified region. Uses pagination to satisfy page size condition.
func (cmd *OffersUpdatesCommand) fetchOffers(errCh chan<- error, regionCh <-chan int64) <-chan api.Offer {
	log.Printf("[DEBUG] Fetching orders..")

	offersCh := make(chan api.Offer)

	go func() {
		defer close(offersCh)

		var offersWg sync.WaitGroup

		for region := range regionCh {
			region := region

			offersWg.Add(1)

			go func() {
				defer offersWg.Done()

				cmd.fetchRegionOffers(errCh, region, offersCh)
			}()
		}

		offersWg.Wait()
	}()
	return offersCh
}

// fetchRegionOffers fetches offers for provided region id
func (cmd *OffersUpdatesCommand) fetchRegionOffers(errCh chan<- error, region int64, offersCh chan<- api.Offer) {
	log.Printf("[DEBUG] Fetching orders for region %v..", region)

	offersPage, err := cmd.PrimaryMarketAPI.GetOffers(
		api.PageableOffersRequest{
			Region:   region,
			Sort:     api.PageableOffersRequestSortCreatedDate,
			PageSize: 50,
		})

	for ok := true; ok; ok = err != nil || len(offersPage.Results) > 0 {

		if err != nil {
			errCh <- err
			return
		}

		log.Printf("[DEBUG] Fetched %v chunk of offers with size %v", offersPage.Page, offersPage.PageSize)

		for _, offer := range offersPage.Results {
			offersCh <- offer
		}

		if offersPage.Next == "" {
			return
		}

		offersPage, err = cmd.PrimaryMarketAPI.GetOffersNextPage(*offersPage)
	}
}

// mapApiOffers maps to domain struct
func (cmd *OffersUpdatesCommand) mapApiOffers(_ chan<- error, apiOfferCh <-chan api.Offer) chan store.Offer {
	log.Printf("[DEBUG] Mapping orders..")

	storeOfferCh := make(chan store.Offer)
	go func() {
		defer close(storeOfferCh)

		for apiOffer := range apiOfferCh {

			log.Printf("[DEBUG] Creating store offer for id %v..", apiOffer.Id)
			storeOffer := store.Offer{
				Id:            apiOffer.Id,
				Slug:          apiOffer.Slug,
				Name:          apiOffer.Name,
				VendorSlug:    apiOffer.Vendor.Slug,
				MainImageLink: apiOffer.MainImage.Image375x211,
				Link:          api.OfferUrl(cmd.PrimaryMarketURL, apiOffer),
				ImportedAt:    cmd.Clock.Now(),
				RegionName:    apiOffer.Region.FullName,
				PriceMin:      apiOffer.Stats.RangesPriceMin,
				PriceMax:      apiOffer.Stats.RangesPriceMax,
				AreaMin:       apiOffer.Stats.RangesAreaMin,
				AreaMax:       apiOffer.Stats.RangesAreaMax,
			}

			storeOfferCh <- storeOffer
		}
	}()
	return storeOfferCh
}

// orchestrateOffers filters already processed offers using offer store, compares prices and redirects
func (cmd *OffersUpdatesCommand) orchestrateOffers(errCh chan<- error, apiOfferCh <-chan store.Offer) (<-chan store.Offer, <-chan store.Offer, <-chan store.Offer) {
	log.Printf("[DEBUG] Filtering orders..")

	newOffersCh := make(chan store.Offer)
	priceRiseCh := make(chan store.Offer)
	priceDropCh := make(chan store.Offer)
	go func() {
		defer func() {
			close(newOffersCh)
			close(priceRiseCh)
			close(priceDropCh)
		}()

		for offer := range apiOfferCh {

			existing, err := cmd.OfferStore.Get(offer.Id)
			if err != nil {
				if _, ok := err.(file.NoPathError); ok {
					log.Printf("[DEBUG] Message id %v does not exist..", offer.Id)
					newOffersCh <- offer
					continue
				}
				errCh <- err
				continue
			}

			diff := existing.CompareAveragePrices(offer)

			if diff < 0 {
				priceRiseCh <- offer
			}
			if diff > 0 {
				priceDropCh <- offer
			}
		}
	}()
	return newOffersCh, priceRiseCh, priceDropCh
}

// writeNewOffers writes an information about newly processed offers
func (cmd *OffersUpdatesCommand) writeNewOffers(errCh chan<- error, offerCh <-chan store.Offer) <-chan store.Offer {
	log.Printf("[DEBUG] Notifying orders updates..")

	notifiedOfferCh := make(chan store.Offer)
	go func() {
		defer close(notifiedOfferCh)

		for offer := range offerCh {

			b := make([]byte, 0)
			if len(offer.MainImageLink) > 0 {

				log.Printf("[DEBUG] Getting main image for offer id %v..", offer.Id)

				imgResp, err := cmd.HttpClient.Get(offer.MainImageLink)
				if err != nil {
					errCh <- err
					continue
				}

				b, err = ioutil.ReadAll(imgResp.Body)
				if err != nil {
					errCh <- err
					continue
				}
			}

			log.Printf("[DEBUG] Creating a notification for offer id %v..", offer.Id)

			txt := "" +
				"🏡" + offer.Name + "\n" +
				"📍 " + offer.RegionName + "\n" +
				"📏 " + strconv.Itoa(offer.AreaMin) + "-" + strconv.Itoa(offer.AreaMax) + "\n"
			if offer.PriceMin > 0 || offer.PriceMax > 0 {
				txt = txt + "🙀 " + strconv.FormatInt(offer.PriceMin, 10) + "-" + strconv.FormatInt(offer.PriceMax, 10) + "\n"
			}
			txt = txt + "\n" +
				"➡️ " + offer.Link

			err := cmd.OfferWriter.Write(writer.Message{
				Title: offer.MainImageLink,
				Image: b,
				Text:  txt,
			})
			if err != nil {
				errCh <- err
				continue
			}
			notifiedOfferCh <- offer
		}
	}()
	return notifiedOfferCh
}

func (cmd *OffersUpdatesCommand) writeOffersPriceRise(errCh chan<- error, offerCh <-chan store.Offer) <-chan store.Offer {
	return cmd.writeOffersPriceChange(errCh, offerCh, "↗️")
}

func (cmd *OffersUpdatesCommand) writeOffersPriceDrop(errCh chan<- error, offerCh <-chan store.Offer) <-chan store.Offer {
	return cmd.writeOffersPriceChange(errCh, offerCh, "↘️")
}

// writeNewOffers writes an information about newly processed offers
func (cmd *OffersUpdatesCommand) writeOffersPriceChange(errCh chan<- error, offerCh <-chan store.Offer, change string) <-chan store.Offer {
	log.Printf("[DEBUG] Notifying orders updates..")

	notifiedOfferCh := make(chan store.Offer)
	go func() {
		defer close(notifiedOfferCh)

		for offer := range offerCh {

			log.Printf("[DEBUG] Creating a notification for offer id %v..", offer.Id)

			err := cmd.OfferWriter.Write(writer.Message{
				Image: make([]byte, 0),
				Text: "" +
					"➡️ " + offer.Link + "\n" +
					"\n" +
					change + " " + strconv.FormatInt(offer.PriceMin, 10) + "-" + strconv.FormatInt(offer.PriceMax, 10),
			})
			if err != nil {
				errCh <- err
				continue
			}
			notifiedOfferCh <- offer
		}
	}()
	return notifiedOfferCh
}

// persistOffers persists offer processing information
func (cmd *OffersUpdatesCommand) persistOffers(doneCh chan<- bool, errCh chan<- error, offerCh <-chan store.Offer) {
	log.Printf("[DEBUG] Persisting orders..")

	go func() {
		for offer := range offerCh {

			log.Printf("[DEBUG] Persisting store offer for id %v..", offer.Id)

			err := cmd.OfferStore.Save(offer)
			if err != nil {
				errCh <- err
			}
		}

		doneCh <- true
	}()
}

func merge(cs ...<-chan store.Offer) <-chan store.Offer {
	var wg sync.WaitGroup
	out := make(chan store.Offer)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan store.Offer) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
