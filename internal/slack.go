// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package scuttle

import (
	"fmt"

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

func (w *Scuttle) notifySlack(action Notification, node string) {
	msg := &slack.WebhookMessage{}

	switch action {
	case Uncordon:
		msg.Text = fmt.Sprintf(":white_check_mark: Uncordon node `%s`", node)
	case TermNotice:
		msg.Text = fmt.Sprintf(":stopwatch: Detected spot termination notice for `%s`", node)
	case Shutdown:
		msg.Text = fmt.Sprintf(":octagonal_sign: Detected shutdown of `%s`", node)
	case Drain:
		msg.Text = fmt.Sprintf(":droplet: Draining node `%s`", node)
	case Delete:
		msg.Text = fmt.Sprintf(":headstone: Deleting node `%s`", node)
	}

	err := slack.PostWebhook(w.config.Webhook, msg)
	if err != nil {
		w.log.Errorf("error notifying Slack webhook url: %v", err)
	}
}
