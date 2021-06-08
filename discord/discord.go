package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/codingpot/alertmanager-discord/alertman"
	"github.com/sirupsen/logrus"
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
		DO := DiscordOut{}

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
			DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
		}

		for _, alert := range alerts {
			realname := alert.Labels["instance"]
			if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
				realname = alert.Labels["exported_instance"]
			}

			RichEmbed.Fields = append(RichEmbed.Fields, DiscordEmbedField{
				Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
				Value: alert.Annotations.Description,
			})
		}

		DO.Embeds = []DiscordEmbed{RichEmbed}

		DOD, _ := json.Marshal(DO)
		http.Post(discordWebhookURL, "application/json", bytes.NewReader(DOD))
	}
}

func SendRawPromAlertWarn(discordWebhookURL string) {
	badString := `This program is suppose to be fed by alertmanager.` + "\n" +
		`It is not a replacement for alertmanager, it is a ` + "\n" +
		`webhook target for it. Please read the README.md  ` + "\n" +
		`for guidance on how to configure it for alertmanager` + "\n" +
		`or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`

	logrus.Print(`/!\ -- You have misconfigured this software -- /!\`)
	logrus.Print(`--- --                                      -- ---`)
	logrus.Print(badString)

	DO := DiscordOut{
		Content: "",
		Embeds: []DiscordEmbed{
			{
				Title:       "You have misconfigured this software",
				Description: badString,
				Color:       ColorGrey,
				Fields:      []DiscordEmbedField{},
			},
		},
	}

	DOD, _ := json.Marshal(DO)
	_, err := http.Post(discordWebhookURL, "application/json", bytes.NewReader(DOD))
	if err != nil {
		logrus.Printf("failed to sned the alert")
		return
	}
}
