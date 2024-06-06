// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	sctl "github.com/poseidon/scuttle/internal"
)

var (
	// version provided by compile time -ldflags
	version = "was not built properly"
	// logger defaults to info logging
	log = logrus.New()
)

func main() {
	flags := struct {
		nodeName string
		channel  string
		token    string
		webhook  string
		platform string
		uncordon bool
		drain    bool
		delete   bool
		logLevel string
		version  bool
		help     bool
	}{}

	flag.StringVar(&flags.nodeName, "node", "", "Kubernetes node name (defaults to $HOSTNAME)")
	flag.StringVar(&flags.platform, "platform", "none", "Set platform (none, aws, azure) to poll termination notices")
	flag.BoolVar(&flags.uncordon, "uncordon", true, "Enabling uncordoning node on start")
	flag.BoolVar(&flags.drain, "drain", true, "Enabling draining node on stop")
	flag.BoolVar(&flags.delete, "delete", true, "Enable deleting node on stop")
	flag.StringVar(&flags.logLevel, "log-level", "info", "Set the logging level")

	// Slack
	flag.StringVar(&flags.channel, "channel-id", "", "Slack Channel ID (e.g. C0FAKEFAKE)")
	flag.StringVar(&flags.token, "token", "", "Slack App token")
	flag.StringVar(&flags.webhook, "webhook", "", "Slack or Google Webhook URL (e.g. \"https://hooks.slack.com\" or \"https://chat.googleapis.com/v1/spaces/\")")

	// subcommands
	flag.BoolVar(&flags.version, "version", false, "Print version and exit")
	flag.BoolVar(&flags.help, "help", false, "Print usage and exit")

	// parse command line arguments
	flag.Parse()

	if flags.version {
		fmt.Println(version)
		return
	}

	if flags.help {
		flag.Usage()
		return
	}

	// logger
	lvl, err := logrus.ParseLevel(flags.logLevel)
	if err != nil {
		log.Fatalf("invalid log-level: %v", err)
	}
	log.Level = lvl

	// allow poll loop to be interrupted
	// buffer to prevent missing a signal if sent before we're receiving
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ctx, stop := context.WithCancel(context.Background())

	// watch for interrupts
	go func() {
		log.Info("main: watch for interrupt signals")
		<-sig

		log.Info("main: detected interrupt")
		// cancel outgoing requests, trigger drain/delete
		stop()
	}()

	// termination watcher
	scuttle, err := sctl.New(&sctl.Config{
		Logger:         log,
		NodeName:       flags.nodeName,
		Platform:       flags.platform,
		ShouldUncordon: flags.uncordon,
		ShouldDrain:    flags.drain,
		ShouldDelete:   flags.delete,
		// Slack
		Channel: flags.channel,
		Token:   flags.token,
		// Slack or gChat
		Webhook: flags.webhook,
	})
	if err != nil {
		log.Fatalf("main: scuttle New error: %v", err)
	}

	// watch for termination
	log.Infof("main: starting scuttle")
	err = scuttle.Run(ctx)
	if err != nil {
		log.Fatalf("main: Run error: %v", err)
	}
	log.Info("done")
}
