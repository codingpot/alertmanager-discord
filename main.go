package main

import (
	"flag"
	"fmt"
	"github.com/codingpot/alertmanager-discord/internal"
	log "github.com/sirupsen/logrus"

	"net/http"
	"os"
)

const defaultPort = "4000"

var (
	discordWebhookURL = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	port              = flag.String("port", os.Getenv("PORT"), "Port to listen on")
)

func main() {
	flag.Parse()
	internal.ValidateWebhookURL(*discordWebhookURL)

	if *port == "" {
		*port = defaultPort
	}

	log.WithField("port", *port).Print("Starting")

	r := internal.NewRouter(*discordWebhookURL)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), r); err != nil {
		log.WithError(err).Panicf("failed to listen")
	}
}
