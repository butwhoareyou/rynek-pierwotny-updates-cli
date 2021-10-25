package cmd

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	log "github.com/go-pkgz/lgr"
	"go.uber.org/multierr"
	"sync"
)

type OffersUpdatesCommand struct {
	PropertiesRequest struct {
		Regions []int64 `short:"r" long:"regions" env:"REGIONS" description:"offer regions"`
	} `group:"request" namespace:"request" env-namespace:"REQUEST"`
	CommonOpts
}

func (cmd *OffersUpdatesCommand) Execute(_ []string) error {
	log.Printf("[DEBUG] Executing offers updates command..")

	resetEnv("TELEGRAM_CHAT_ID", "TELEGRAM_TOKEN")

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
		cmd.persistOffers(doneCh, errCh,
			cmd.notifyOffersUpdates(errCh,
				cmd.mapApiOffers(errCh,
					cmd.filterNewOffers(errCh,
						cmd.fetchOffers(errCh,
							cmd.streamRegions())))))
	}()
}

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

// filterNewOffers filters already processed offers using offer store
func (cmd *OffersUpdatesCommand) filterNewOffers(errCh chan<- error, apiOfferCh <-chan api.Offer) <-chan api.Offer {
	log.Printf("[DEBUG] Filtering orders..")

	filteredOffersCh := make(chan api.Offer)
	go func() {
		defer close(filteredOffersCh)

		for offer := range apiOfferCh {
			exists, err := cmd.OfferStore.Exists(offer.Id)
			if err != nil {
				errCh <- err
				continue
			}

			if !exists {
				log.Printf("[DEBUG] Offer id %v does not exist..", offer.Id)
				filteredOffersCh <- offer
			}
		}
	}()
	return filteredOffersCh
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

// notifyOffersUpdates sends a notification about newly processed offers
func (cmd *OffersUpdatesCommand) notifyOffersUpdates(errCh chan<- error, offerCh <-chan store.Offer) <-chan store.Offer {
	log.Printf("[DEBUG] Notifying orders updates..")

	notifiedOfferCh := make(chan store.Offer)
	go func() {
		defer close(notifiedOfferCh)

		for offer := range offerCh {
			log.Printf("[DEBUG] Creating notification for offer id %v..", offer.Id)
			err := cmd.OfferNotifier.Notify(offer)
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

			err := cmd.OfferStore.Create(offer)
			if err != nil {
				errCh <- err
			}
		}

		doneCh <- true
	}()
}
