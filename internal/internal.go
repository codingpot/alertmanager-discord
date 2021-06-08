package internal

import (
	"encoding/json"
	"github.com/codingpot/alertmanager-discord/alertman"
	"github.com/codingpot/alertmanager-discord/discord"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

func ValidateWebhookURL(webhookURL string) {
	if webhookURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}
	_, err := url.Parse(webhookURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(webhookURL)); !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid.")
	}
}

func NewRouter(discordWebhookURL string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Heartbeat("/healthcheck"))

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var amo alertman.AlertManOut
		err = json.Unmarshal(b, &amo)
		if err != nil {
			if alertman.IsRawPromAlert(b) {
				discord.SendRawPromAlertWarn(discordWebhookURL)
				return
			}

			if len(b) > 1024 {
				logrus.Printf("Failed to unpack inbound alert request - %s...", string(b[:1023]))

			} else {
				logrus.Printf("Failed to unpack inbound alert request - %s", string(b))
			}

			return
		}

		discord.SendWebhook(&amo, discordWebhookURL)
	})

	return r
}
