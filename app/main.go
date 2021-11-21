package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	as3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/cmd"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/file"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/mock"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/s3"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/util"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/writer/telegram"
	log "github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/umputun/go-flags"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

type Opts struct {
	OffersUpdates cmd.OffersUpdatesCommand `command:"offers-updates"`

	PrimaryMarketPLURL    string `long:"url" env:"URL" required:"true" description:"RynekPierwotny.pl url"`
	PrimaryMarketAPIPLURL string `long:"api-url" env:"API_URL" required:"true" description:"RynekPierwotny.pl api url"`

	FileSystem struct {
		StorePath string `long:"store-path" env:"STORE_PATH" description:"Store path to directory with execution state"`
	} `group:"fs" namespace:"fs" env-namespace:"FS"`

	AWS struct {
		S3 struct {
			Bucket string `long:"bucket" env:"BUCKET" description:"Execution state store bucket"`
		} `group:"s3" namespace:"s3" env-namespace:"S3"`
		Endpoint string `long:"endpoint" env:"ENDPOINT" description:"Use when request is going to any s3-like storage"`
		Region   string `long:"region" env:"REGION" description:"AWS region"`
	} `group:"aws" namespace:"aws" env-namespace:"AWS"`

	Telegram struct {
		ChatId int64  `long:"chat-id" env:"CHAT_ID" description:"Chat id notifications will be sent to"`
		Token  string `long:"token" env:"TOKEN" description:"Token will be used to send notifications"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Debug bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		setupLog(opts.Debug)

		eng, err := setupEngine(opts)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
			return err
		}

		offerStore, err := setupOfferStore(eng)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
			return err
		}

		botApi, err := setupBotApi(opts)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
			return err
		}

		httpClient := http.Client{}
		offerNotifier := setupOfferWriter(opts, botApi)

		c := command.(cmd.CommonCommander)
		c.SetCommon(cmd.CommonOpts{
			PrimaryMarketURL: opts.PrimaryMarketPLURL,
			PrimaryMarketAPI: api.NewHttpApi(opts.PrimaryMarketAPIPLURL),
			OfferStore:       *offerStore,
			OfferWriter:      *offerNotifier,
			Clock:            util.EagerClock{},
			HttpClient:       httpClient,
		})
		err = c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}
		return err
	}
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func setupBotApi(opts Opts) (*tgbotapi.BotAPI, error) {
	if opts.Telegram.Token != "" {
		log.Print("[DEBUG] Telegram token provided.")
		return tgbotapi.NewBotAPI(opts.Telegram.Token)
	}
	// ignore lack of tg bot api - it is optional
	return nil, nil
}

func setupOfferWriter(opts Opts, botAPI *tgbotapi.BotAPI) *writer.MessageWriter {
	var w writer.MessageWriter
	w = &writer.LogWriter{}
	if botAPI != nil && opts.Telegram.ChatId != 0 {
		log.Print("[DEBUG] Telegram writer initialized.")
		w = telegram.NewWriter(opts.Telegram.ChatId, botAPI)
	}
	return &w
}

func setupEngine(opts Opts) (engine.Engine, error) {
	if opts.AWS.S3.Bucket != "" && opts.AWS.Region != "" {
		endpoint := stringOrNil(opts.AWS.Endpoint)
		sess, err := session.NewSession(&aws.Config{
			Region:           aws.String(opts.AWS.Region),
			Endpoint:         endpoint,
			DisableSSL:       aws.Bool(endpoint != nil),
			S3ForcePathStyle: aws.Bool(endpoint != nil),
		})
		if err != nil {
			return nil, err
		}
		svc := as3.New(sess)
		eng, err := s3.NewEngine(opts.AWS.S3.Bucket, svc)
		return eng, err
	}
	if opts.FileSystem.StorePath != "" {
		eng, err := file.NewSystemEngine(opts.FileSystem.StorePath)
		return eng, err
	}
	eng := mock.NewEngine()
	return eng, nil
}

func setupOfferStore(engine engine.Engine) (*store.OfferStore, error) {
	fs := store.NewOfferFileStore(engine)
	return &fs, nil
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}

// getDump reads runtime stack and returns as a string
func getDump() string {
	maxSize := 5 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

// nolint:gochecknoinits // can't avoid it in this place
func init() {
	// catch SIGQUIT and print stack traces
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}

func stringOrNil(s string) *string {
	if s != "" {
		return &s
	}
	return nil
}
