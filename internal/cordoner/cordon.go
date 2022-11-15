// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package cordon

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// Cordoner manages cordoning nodes.
type Cordoner interface {
	// Cordon marks a Kubernetes Node as unschedulable.
	Cordon(ctx context.Context, node string) error
	// Uncordon marks a Kubernetes Node as schedulable.
	Uncordon(ctx context.Context, node string) error
}

// Config configures a Cordoner.
type Config struct {
	Client kubernetes.Interface
	Logger *logrus.Logger
}

// New returns a new Drainer.
func New(config *Config) Cordoner {
	return &cordoner{
		client: config.Client,
		log:    config.Logger,
	}
}

// cordoner is a Kubernetes node cordoner.
type cordoner struct {
	client kubernetes.Interface
	log    *logrus.Logger
}

// Cordon marks a Kubernetes Node as unschedulable.
func (d *cordoner) Cordon(ctx context.Context, node string) error {
	d.log.WithField("node", node).Info("drainer: cordoning node")
	return d.setUnschedulable(ctx, node, true)
}

// Uncordon marks a Kubernetes Node as schedulable.
func (d *cordoner) Uncordon(ctx context.Context, node string) error {
	d.log.WithField("node", node).Info("drainer: uncordoning node")
	return d.setUnschedulable(ctx, node, false)
}

// setUnschedulable updates a Node's spec to mark it unschedulable or not.
func (d *cordoner) setUnschedulable(ctx context.Context, node string, unschedule bool) error {
	patch := []byte(fmt.Sprintf("{\"spec\":{\"unschedulable\":%t}}", unschedule))
	_, err := d.client.CoreV1().Nodes().Patch(ctx, node, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}
