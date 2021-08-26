package cmd

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/notification"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/util"
	log "github.com/go-pkgz/lgr"
	"os"
)

type CommonCommander interface {
	SetCommon(commonOpts CommonOpts)

	Execute(args []string) error
}

type CommonOpts struct {
	PrimaryMarketURL string
	PrimaryMarketAPI api.Api
	OfferStore       store.OfferStore
	OfferNotifier    notification.OfferNotifier
	Clock            util.Clock
}

func (c *CommonOpts) SetCommon(commonOpts CommonOpts) {
	c.PrimaryMarketURL = commonOpts.PrimaryMarketURL
	c.PrimaryMarketAPI = commonOpts.PrimaryMarketAPI
	c.OfferStore = commonOpts.OfferStore
	c.OfferNotifier = commonOpts.OfferNotifier
	c.Clock = commonOpts.Clock
}

// resetEnv clears sensitive env vars
func resetEnv(envs ...string) {
	for _, env := range envs {
		if err := os.Unsetenv(env); err != nil {
			log.Printf("[WARN] can't unset env %s, %s", env, err)
		}
	}
}
