package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/getsentry/raven-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"googlemaps.github.io/maps"
	"gopkg.in/telegram-bot-api.v4"
	"rsc.io/letsencrypt"
)

var (
	SecretsPath          string
	LetsencryptCachePath string
	Hostname             string

	DB *sqlx.DB

	BotAPI *tgbotapi.BotAPI
	Maps   *maps.Client

	Secrets struct {
		BotToken                 string `json:"bot_token"`
		PostgresConnectionString string `json:"postgres_connection_string"`
		MapsAPIKey               string `json:"maps_api_key"`
		SentryDSN                string `json:"sentry_dsn"`
	}
)

func Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return BotAPI.Send(c)
}

func readSecrets() error {
	secretsFile, err := os.Open(SecretsPath)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(secretsFile)
	err = decoder.Decode(&Secrets)
	if err != nil {
		return err
	}
	if Secrets.SentryDSN != "" {
		raven.SetDSN(Secrets.SentryDSN)
	}
	if Secrets.BotToken == "" {
		return errors.New("bot_token is missing in secrets")
	}
	BotAPI, err = tgbotapi.NewBotAPI(Secrets.BotToken)
	if err != nil {
		return err
	}
	if Secrets.MapsAPIKey == "" {
		return errors.New("maps_api_key is missing in secrets")
	}
	Maps, err = maps.NewClient(maps.WithAPIKey(Secrets.MapsAPIKey))
	if err != nil {
		return err
	}
	if Secrets.PostgresConnectionString == "" {
		return errors.New("postgres_connection_string is missing in secrets")
	}
	DB, err = sqlx.Open("postgres", Secrets.PostgresConnectionString)
	if err != nil {
		return err
	}
	_, err = DB.Query("SET search_path TO srsbot")
	return err
}

//go:generate file2const --package=main site/index.html:Site site.go
func createHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/telegram/webhook/"+Secrets.BotToken, raven.RecoveryHandler(handleTelegramWebhook))
	mux.HandleFunc("/telegram/register_webhook/"+Secrets.BotToken, func(w http.ResponseWriter, r *http.Request) {
		_, err := BotAPI.SetWebhook(tgbotapi.NewWebhook(fmt.Sprintf("https://%s/telegram/webhook/%s", Hostname, Secrets.BotToken)))
		if err == nil {
			http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(Site))
	})
	return mux
}

func main() {
	flag.StringVar(&SecretsPath, "secrets", "", "Path to secrets file")
	flag.StringVar(&LetsencryptCachePath, "letsencrypt-cache", "", "Path to Let's Encrypt cache file")
	flag.StringVar(&Hostname, "hostname", "", "Hostname to register webhook with")
	flag.Parse()

	if SecretsPath == "" || LetsencryptCachePath == "" || Hostname == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := readSecrets(); err != nil {
		log.Fatal(err)
	}

	var m letsencrypt.Manager
	if err := m.CacheFile(LetsencryptCachePath); err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(letsencrypt.RedirectHTTP),
	}

	httpsServer := &http.Server{
		Addr:    ":8443",
		Handler: createHandler(),
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
	}

	go Poller()
	if err := gracehttp.Serve(httpServer, httpsServer); err != nil {
		log.Fatal(err)
	}
}
