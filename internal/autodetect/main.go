// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package autodetect is for auto-detecting traits from the environment (platform, APIs, ...).
package autodetect

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// AutoDetect provides an assortment of routines that auto-detect traits based on the runtime.
type AutoDetect struct {
	logger logr.Logger
	dcl    discovery.DiscoveryInterface
}

// New creates a new auto-detection worker, using the given client when talking to the current cluster.
func New(logger logr.Logger, restConfig *rest.Config) (*AutoDetect, error) {
	dcl, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		// it's pretty much impossible to get into this problem, as most of the
		// code branches from the previous call just won't fail at all,
		// but let's handle this error anyway...
		return nil, err
	}

	return &AutoDetect{
		logger: logger,
		dcl:    dcl,
	}, nil
}

// Distro returns the detected distro this operator is running on. Possible values: Kubernetes, OpenShift.
func (a *AutoDetect) Distro() Distro {
	apiList, err := a.dcl.ServerGroups()
	if err != nil {
		a.logger.Error(err, "unable to detect platform")
		return UnknownDistro
	}

	apiGroups := apiList.Groups
	for i := 0; i < len(apiGroups); i++ {
		if apiGroups[i].Name == "route.openshift.io" {
			return OpenShiftDistro
		}
	}

	return KubernetesDistro
}
