package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codingpot/alertmanager-discord/alertman"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// Discord color values
const (
	ColorRed   = 0x992D22
	ColorGreen = 0x2ECC71
	ColorGrey  = 0x95A5A6
)

type DiscordOut struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields"`
}

type DiscordEmbedField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func SendWebhook(amo *alertman.AlertManOut, discordWebhookURL string) {
	groupedAlerts := make(map[string][]alertman.AlertManAlert)

	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	for status, alerts := range groupedAlerts {
		var discordOut DiscordOut

		RichEmbed := DiscordEmbed{
			Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
			Description: amo.CommonAnnotations.Summary,
			Color:       ColorGrey,
			Fields:      []DiscordEmbedField{},
		}

		if status == "firing" {
			RichEmbed.Color = ColorRed
		} else if status == "resolved" {
			RichEmbed.Color = ColorGreen
		}

		if amo.CommonAnnotations.Summary != "" {
			discordOut.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
		}

		for _, alert := range alerts {
			realName := alert.Labels["instance"]
			if strings.Contains(realName, "localhost") && alert.Labels["exported_instance"] != "" {
				realName = alert.Labels["exported_instance"]
			}

			RichEmbed.Fields = append(RichEmbed.Fields, DiscordEmbedField{
				Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realName),
				Value: alert.Annotations.Description,
			})
		}

		discordOut.Embeds = []DiscordEmbed{RichEmbed}
		DOD, _ := json.Marshal(discordOut)

		_, err := http.Post(discordWebhookURL, "application/json", bytes.NewReader(DOD))
		if err != nil {
			log.WithError(err).Error("failed to write to webhook")
			return
		}
	}
}

func SendRawPromAlertWarn(discordWebhookURL string) {
	badString := `This program is suppose to be fed by alertmanager.` + "\n" +
		`It is not a replacement for alertmanager, it is a ` + "\n" +
		`webhook target for it. Please read the README.md  ` + "\n" +
		`for guidance on how to configure it for alertmanager` + "\n" +
		`or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`

	log.Println(`/!\ -- You have misconfigured this software -- /!\`)
	log.Println(`--- --                                      -- ---`)
	log.Println(badString)

	discordOutBytes, _ := json.Marshal(DiscordOut{
		Content: "",
		Embeds: []DiscordEmbed{
			{
				Title:       "You have misconfigured this software",
				Description: badString,
				Color:       ColorGrey,
				Fields:      []DiscordEmbedField{},
			},
		},
	})
	_, err := http.Post(discordWebhookURL, "application/json", bytes.NewReader(discordOutBytes))
	if err != nil {
		log.Printf("failed to sned the alert")
		return
	}
}
