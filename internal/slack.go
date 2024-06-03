// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package scuttle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"go.seankhliao.com/gchat"
)

type Notification string

const (
	Uncordon   Notification = "uncordon"
	TermNotice Notification = "term-notice"
	Shutdown   Notification = "shutdown"
	Drain      Notification = "drain"
	Delete     Notification = "delete"
)

func ErrorNotification(err error) Notification {
	return Notification(err.Error())
}

// gchat
func gchatReport(client *gchat.WebhookClient, obj string) {
	// if strings.Contains(obj, "ERROR") {
	// 	return fmt.Sprintf("WebhookClient: Unable to send - %s'", obj)
	// }
	err := client.Post(context.Background(), gchat.WebhookPayload{
		Text: obj,
	})
	if err != nil {
		fmt.Sprintf("WebhookClient: Unable to send - %s'", err)
	}
}

// notifySlack posts a Slack message (and reaction) and returns the message
// timestamp for threading subsequent replies.
// - Slack client mode posts a Slack message or reply (if thread set) and
// reaction
// - Slack webhook mode just posts a simple message only
func (w *Scuttle) notifySlack(action Notification, thread string) string {
	var text, reaction, color string

	switch action {
	case Uncordon:
		color = "good"
		text = fmt.Sprintf("🐣 Uncordon node `%s`", w.hostname)
	case TermNotice:
		color = "warning"
		text = fmt.Sprintf("⏱️ Detected spot termination notice for `%s`", w.hostname)
	case Shutdown:
		color = "warning"
		text = fmt.Sprintf("⚠️ Detected shutdown of `%s`", w.hostname)
	case Drain:
		color = "warning"
		text = fmt.Sprintf("💧 Draining node `%s`", w.hostname)
		reaction = "droplet"
	case Delete:
		color = "warning"
		text = fmt.Sprintf("🪦 Deleting node `%s`", w.hostname)
		reaction = "headstone"
	default:
		color = "danger"
		text = fmt.Sprintf("‼️ %s ‼️", action)
		reaction = "red_circle"
	}

	now := time.Now().Format(time.StampMilli)
	attachment := slack.Attachment{
		Color:  color,
		Text:   text,
		Footer: now,
		Ts:     json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	// Slack App token mode (richer)
	if w.config.Webhook == "" {
		if w.slack != nil {
			if reaction != "" {
				msgRef := slack.NewRefToMessage(w.config.Channel, thread)
				if err := w.slack.AddReaction(reaction, msgRef); err != nil {
					w.log.Errorf("error posting Slack reaction: %v", err)
				}
			}

			opts := []slack.MsgOption{
				slack.MsgOptionAttachments(attachment),
			}
			if thread != "" {
				opts = append(opts, slack.MsgOptionTS(thread))
			}

			_, stamp, err := w.slack.PostMessage(
				w.config.Channel,
				opts...,
			)
			if err != nil {
				w.log.Errorf("error posting Slack message: %v", err)
			}

			return stamp
		}
	}

	// Slack App Webhook mode
	if w.config.Webhook == "slack" {
		msg := &slack.WebhookMessage{
			Attachments: []slack.Attachment{
				attachment,
			},
		}
		err := slack.PostWebhook(w.config.Webhook, msg)
		if err != nil {
			w.log.Errorf("error sending Slack Webhook: %v", err)
		}
	}

	// Google Chat Webhook mode
	var chat *gchat.WebhookClient

	if w.config.Webhook != "" {
		var gchatEndpoint = w.config.Webhook
		fmt.Sprintf(w.config.Webhook)

		if gchatEndpoint != "" {
			chat = &gchat.WebhookClient{
				//Client:   &http.Client,
				Endpoint: gchatEndpoint,
			}
		}

		msg := attachment.Text

		if chat != nil {
			gchatReport(chat, msg)
		}
	}

	// only client mode supports threading
	return ""
}
