package main

import (
	"errors"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/api"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/cmd"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/notification"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/notification/telegram"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/file"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/util"
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

	FileSystemStorePath string `long:"fs-store-path" env:"FILE_SYSTEM_STORE_PATH" required:"true" description:"file system path to directory with execution state"`

	TelegramChatId int64  `long:"telegram-chat-id" env:"TELEGRAM_CHAT_ID" description:"Telegram chat id notifications will be sent to"`
	TelegramToken  string `long:"telegram-token" env:"TELEGRAM_TOKEN" description:"Telegram token will be used to send notifications"`

	Debug bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		setupLog(opts.Debug)

		engine := file.NewSystemEngine()
		offerStore, err := setupOfferStore(opts, &engine)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}

		botApi, err := setupBotApi(opts)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}

		httpClient := http.Client{}
		offerNotifier := setupOfferNotifier(opts, botApi, httpClient)

		c := command.(cmd.CommonCommander)
		c.SetCommon(cmd.CommonOpts{
			PrimaryMarketURL: opts.PrimaryMarketPLURL,
			PrimaryMarketAPI: api.NewHttpApi(opts.PrimaryMarketAPIPLURL),
			OfferStore:       *offerStore,
			OfferNotifier:    *offerNotifier,
			Clock:            util.EagerClock{},
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
	if opts.TelegramToken != "" {
		log.Print("[DEBUG] Telegram token provided.")
		return tgbotapi.NewBotAPI(opts.TelegramToken)
	}
	// ignore lack of tg bot api - it is optional
	return nil, nil
}

func setupOfferNotifier(opts Opts, botAPI *tgbotapi.BotAPI, httpClient http.Client) *notification.OfferNotifier {
	var notifier notification.OfferNotifier
	notifier = &notification.LogNotifier{}
	if botAPI != nil && opts.TelegramChatId != 0 {
		log.Print("[DEBUG] Telegram notifier initialized.")
		notifier = telegram.NewNotifier(opts.TelegramChatId, botAPI, httpClient)
	}
	return &notifier
}

func setupOfferStore(opts Opts, engine *file.SystemEngine) (*store.OfferStore, error) {
	if opts.FileSystemStorePath != "" {
		fs, err := store.NewOfferFileStore(opts.FileSystemStorePath+"/offer-updates", engine)
		return &fs, err
	}
	return nil, errors.New("unable to initialize offer store")
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
