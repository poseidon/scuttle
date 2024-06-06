// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package scuttle

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

func (w *Scuttle) notify(action Notification, thread string) string {

	if w.gchat != nil {
		w.notifyGchat(action)
	} else {
		w.notifySlack(action, w.lastThread)
	}

	return ""
}
