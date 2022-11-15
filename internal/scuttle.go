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
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubectl "k8s.io/kubectl/pkg/drain"

	cordon "github.com/poseidon/scuttle/internal/cordoner"
)

const (
	awsMetadata   = "http://169.254.169.254/latest/meta-data/spot/instance-action"
	azureMetadata = "http://169.254.169.254/metadata/scheduledevents?api-version=2019-01-01"
)

// Config configures a Scuttle
type Config struct {
	Logger         *logrus.Logger
	Platform       string
	ShouldUncordon bool
	ShouldDrain    bool
	ShouldDelete   bool
}

// Scuttle watches for termination notices and performs Kubelet teardown actions.
type Scuttle struct {
	hostname   string
	config     *Config
	log        *logrus.Logger
	client     *http.Client
	kubeClient kubernetes.Interface
}

// New creates a new Scuttle.
func New(config *Config) (*Scuttle, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("scuttle: logger must be non-nil")
	}

	// Use HOSTNAME to identify Kubelet node
	hostname := os.Getenv("HOSTNAME")

	// Kubernetes client from kubeconfig or service account (in-cluster)
	kubeconfigPath := os.Getenv("KUBECONFIG")
	kubeClient, err := newKubeClient(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("scuttle: error creating Kubernetes client: %v", err)
	}

	return &Scuttle{
		hostname: hostname,
		config:   config,
		log:      config.Logger,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		kubeClient: kubeClient,
	}, nil
}

// Run runs the spot termination watch loop.
func (w *Scuttle) Run(ctx context.Context) error {
	fields := logrus.Fields{
		"hostname": w.hostname,
	}
	w.log.WithFields(fields).Info("start scuttle")

	// poll spot termination notices
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// best-effort, we need to continue even on error
	err := w.start(ctx)
	if err != nil {
		w.log.WithFields(fields).Errorf("scuttle: uncordon error: %v", err)
	}

	// context for stop/teardown is independent
	stopCtx := context.Background()

	for {
		select {
		case <-ctx.Done():
			w.log.WithFields(fields).Info("scuttle: stopping...")
			return w.stop(stopCtx)
		case <-ticker.C:
			w.log.WithFields(fields).Debug("scuttle: tick...")
			if w.pendingShutdown(ctx) {
				return w.stop(stopCtx)
			}
		}
	}
}

// start optionally runs uncordon.
func (w *Scuttle) start(ctx context.Context) error {
	fields := logrus.Fields{
		"hostname": w.hostname,
	}

	if w.config.ShouldUncordon {
		w.log.WithFields(fields).Info("scuttle: uncordon node")
		cordoner := cordon.New(&cordon.Config{
			Client: w.kubeClient,
			Logger: w.log,
		})
		return cordoner.Uncordon(ctx, w.hostname)
	}

	w.log.WithFields(fields).Info("scuttle: SKIP uncordon node")
	return nil
}

// stop optionally drains and/or deletes a node.
func (w *Scuttle) stop(ctx context.Context) error {
	fields := logrus.Fields{
		"hostname": w.hostname,
	}

	draino := &kubectl.Helper{
		Client:              w.kubeClient,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		GracePeriodSeconds:  -1,
		// upstream drain Helper logs to an Out and ErrOut io.Writer
		// https://pkg.go.dev/k8s.io/kubectl@v0.25.4/pkg/drain#Helper
		Out:    w.log.Writer(),
		ErrOut: w.log.WriterLevel(logrus.WarnLevel),
	}

	cordoner := cordon.New(&cordon.Config{
		Client: w.kubeClient,
		Logger: w.log,
	})

	// optionally drain to evict pods on the node
	if w.config.ShouldDrain {
		w.log.WithFields(fields).Info("scuttle: cordoning node")
		err := cordoner.Cordon(ctx, w.hostname)
		// best-effort, we need to continue even on error
		if err != nil {
			w.log.WithFields(fields).Errorf("scuttle: cordon error: %v", err)
		}

		w.log.WithFields(fields).Info("scuttle: draining node")
		err = kubectl.RunNodeDrain(draino, w.hostname)
		// best-effort, we need to continue even on error
		if err != nil {
			w.log.WithFields(fields).Errorf("scuttle: drain error: %v", err)
		}
	} else {
		w.log.WithFields(fields).Info("scuttle: SKIP drain node")
	}

	// optionally delete the node from the cluster
	if w.config.ShouldDelete {
		w.log.WithFields(fields).Info("scuttle: deleting node")
		err := w.kubeClient.CoreV1().Nodes().Delete(ctx, w.hostname, v1.DeleteOptions{})
		// best-effort, we need to continue even on error
		if err != nil {
			w.log.WithFields(fields).Errorf("scuttle: delete error: %v", err)
		}
	} else {
		w.log.WithFields(fields).Info("scuttle: SKIP delete node")
	}

	return nil
}

// Check metadata for pending termination notices
func (w *Scuttle) pendingShutdown(ctx context.Context) bool {
	fields := logrus.Fields{
		"hostname": w.hostname,
	}

	endpoint := ""
	switch w.config.Platform {
	case "aws":
		endpoint = awsMetadata
	case "azure":
		endpoint = azureMetadata
	default:
		w.log.Warn("scuttle: SKIP checking cloud metadata for pending terminations")
		return false
	}

	w.log.WithFields(fields).Debug("scuttle: check for termination notices")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		w.log.WithFields(fields).Errorf("scuttle: invalid request: %v", err)
		return false
	}

	req.Header.Set("Metadata", "true") // make Azure happy, AWS doesn't mind
	resp, err := w.client.Do(req)
	if err != nil {
		w.log.WithFields(fields).Errorf("scuttle: metadata error: %v", err)
		return false
	}

	switch resp.StatusCode {
	case 200:
		w.log.WithFields(fields).Info("scuttle: Spot Instance interruption notice!")
		return true
	default:
		w.log.WithFields(fields).Debugf("scuttle: metadata status code %d", resp.StatusCode)
		return false
	}
}
