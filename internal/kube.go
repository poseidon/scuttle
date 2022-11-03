// Copyright (C) 2022 Poseidon Labs
// Copyright (C) 2022 Dalton Hubble
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package scuttle

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// newKubeClient creates a Kubernetes client using a kubeconfig at the given
// path or using the Pod service account (i.e. in-cluster).
func newKubeClient(kubePath string) (kubernetes.Interface, error) {
	// Kubernetes REST client config
	config, err := clientcmd.BuildConfigFromFlags("", kubePath)
	if err != nil {
		return nil, fmt.Errorf("scuttle: error getting Kubernetes client config: %v", err)
	}

	// create Kubernetes client
	return kubernetes.NewForConfig(config)
}
