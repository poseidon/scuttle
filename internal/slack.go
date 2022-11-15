// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package scuttle

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

type Notification string

const (
	Uncordon   Notification = "uncordon"
	TermNotice Notification = "term-notice"
	Shutdown   Notification = "shutdown"
	Drain      Notification = "drain"
	Delete     Notification = "delete"
)

// notifySlack posts a Slack message (and reaction) and returns the message
// timestamp for threading subsequent replies.
// - Slack client mode posts a Slack message or reply (if thread set) and
// reaction
// - Slack webhook mode just posts a simple message only
func (w *Scuttle) notifySlack(action Notification, node string, thread string) string {
	var text, reaction, color string

	switch action {
	case Uncordon:
		color = "good"
		text = fmt.Sprintf(":hatched_chick: Uncordon node `%s`", node)
	case TermNotice:
		color = "warning"
		text = fmt.Sprintf(":stopwatch: Detected spot termination notice for `%s`", node)
	case Shutdown:
		color = "warning"
		text = fmt.Sprintf(":warning: Detected shutdown of `%s`", node)
	case Drain:
		color = "warning"
		text = fmt.Sprintf(":droplet: Draining node `%s`", node)
		reaction = "droplet"
	case Delete:
		color = "warning"
		text = fmt.Sprintf(":headstone: Deleting node `%s`", node)
		reaction = "headstone"
	}

	now := time.Now().Format(time.StampMilli)
	attachment := slack.Attachment{
		Color:  color,
		Text:   text,
		Footer: now,
		Ts:     json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	// Slack App token mode (richer)
	if w.slack != nil {
		if reaction != "" {
			msgRef := slack.NewRefToMessage(w.config.Channel, thread)
			if err := w.slack.AddReaction(reaction, msgRef); err != nil {
				w.log.Errorf("error posting Slack reaction: %v", err)
			}
		}

		opts := []slack.MsgOption{
			//slack.MsgOptionText(text, true),
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

	// Slack App Webhook mode
	if w.config.Webhook != "" {
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

	// only client mode supports threading
	return ""
}
