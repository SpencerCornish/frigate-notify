package notifier

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/0x2142/frigate-notify/config"
	"github.com/0x2142/frigate-notify/models"
	"github.com/disgoorg/disgo/discord"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

// SendSlackMessage pushes alert message to Slack via webhook
func SendSlackMessage(event models.Event, snapshot io.Reader, provider notifMeta) {
	profile := config.ConfigData.Alerts.Slack[provider.index]
	status := &config.Internal.Status.Notifications.Slack[provider.index]

	var err error
	var message string
	// Build notification
	if profile.Template != "" {
		message = renderMessage(profile.Template, event, "message", "Slack")
	} else {
		message = renderMessage("markdown", event, "message", "Slack")
	}

	attachment := slack.Attachment{
		Color:         "good",
		Fallback:      "You successfully posted by Incoming Webhook URL!",
		AuthorName:    "slack-go/slack",
		AuthorSubname: "github.com",
		AuthorLink:    "https://github.com/slack-go/slack",
		AuthorIcon:    "https://avatars2.githubusercontent.com/u/652790",
		Text:          "<!channel> All text in Slack uses the same system of escaping: chat messages, direct messages, file comments, etc. :smile:\nSee <https://api.slack.com/docs/message-formatting#linking_to_channels_and_users>",
		Footer:        "slack api",
		FooterIcon:    "https://platform.slack-edge.com/img/default_application_icon.png",
		Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook("YOUR_WEBHOOK_URL_HERE", &msg)
	if err != nil {
		fmt.Println(err)
	}

	title := renderMessage(config.ConfigData.Alerts.General.Title, event, "title", "Discord")
	title = fmt.Sprintf("**%v**\n\n", title)
	message = title + message

	// Send alert & attach snapshot if one was saved
	var msg *discord.Message
	if event.HasSnapshot {
		image := discord.NewFile("snapshot.jpg", "", snapshot)
		embed := discord.NewEmbedBuilder().SetDescription(message).SetTitle(title).SetImage("attachment://snapshot.jpg").SetColor(5793266).Build()
		msg, err = client.CreateMessage(discord.NewWebhookMessageCreateBuilder().SetEmbeds(embed).SetFiles(image).Build())
		log.Trace().
			Str("event_id", event.ID).
			Int("provider_id", provider.index).
			Interface("payload", msg).
			Msg("Send Discord Alert")

	} else {
		embed := discord.NewEmbedBuilder().SetDescription(message).SetTitle(title).SetColor(5793266).Build()
		msg, err = client.CreateMessage(discord.NewWebhookMessageCreateBuilder().SetEmbeds(embed).Build())
		log.Trace().
			Str("event_id", event.ID).
			Int("provider_id", provider.index).
			Interface("payload", msg).
			Msg("Send Discord Alert")
	}
	if err != nil {
		log.Warn().
			Str("event_id", event.ID).
			Str("provider", "Discord").
			Int("provider_id", provider.index).
			Err(err).
			Msg("Unable to send alert")
		status.NotifFailure(err.Error())
	}

	log.Info().
		Str("event_id", event.ID).
		Str("provider", "Discord").
		Int("provider_id", provider.index).
		Msg("Alert sent")
	status.NotifSuccess()
}
