package cmd

import (
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/util"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer"
	log "github.com/go-pkgz/lgr"
	"net/http"
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
	OfferWriter      writer.MessageWriter
	Clock            util.Clock
	HttpClient       http.Client
}

func (c *CommonOpts) SetCommon(commonOpts CommonOpts) {
	c.PrimaryMarketURL = commonOpts.PrimaryMarketURL
	c.PrimaryMarketAPI = commonOpts.PrimaryMarketAPI
	c.OfferStore = commonOpts.OfferStore
	c.OfferWriter = commonOpts.OfferWriter
	c.Clock = commonOpts.Clock
	c.HttpClient = commonOpts.HttpClient
}

// resetEnv clears sensitive env vars
func resetEnv(envs ...string) {
	for _, env := range envs {
		if err := os.Unsetenv(env); err != nil {
			log.Printf("[WARN] can't unset env %s, %s", env, err)
		}
	}
}
