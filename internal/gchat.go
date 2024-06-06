// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package scuttle

import (
	"context"
	"fmt"
	"time"

	"go.seankhliao.com/gchat"
)

// notifyGchat posts a message to google spaces message with a webhook.
func (w *Scuttle) notifyGchat(action Notification) string {
	var text string

	switch action {
	case Uncordon:
		text = fmt.Sprintf("🐣 Uncordon node %s", w.hostname)
	case TermNotice:
		text = fmt.Sprintf("⏱️ Detected spot termination notice for %s", w.hostname)
	case Shutdown:
		text = fmt.Sprintf("⚠️ Detected shutdown of %s", w.hostname)
	case Drain:
		text = fmt.Sprintf("💧 Draining node %s", w.hostname)
	case Delete:
		text = fmt.Sprintf("🪦 Deleting node %s", w.hostname)
	default:
		text = fmt.Sprintf("🛑 %s", action)
	}

	now := time.Now().Format(time.StampMilli)

	var msg = fmt.Sprintf("`%s` `%s`", now, text)

	attachment := &gchat.WebhookPayload{
		Text: msg,
	}

	if attachment != nil {
		err := w.gchat.Post(context.Background(), *attachment)
		if err != nil {
			w.log.Warn("WebhookClient: Unable to send - `%s`", err)
		}
	}

	return ""
}
