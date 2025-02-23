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

package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
)

func TestDefaultAnnotations(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "my-ns",
		},
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Config: "test",
		}},
	}

	// test
	annotations := Annotations(otelcol)

	//verify
	assert.Equal(t, "true", annotations["prometheus.io/scrape"])
	assert.Equal(t, "8888", annotations["prometheus.io/port"])
	assert.Equal(t, "/metrics", annotations["prometheus.io/path"])
	assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", annotations["splunk-otel-operator-config/sha256"])
}

func TestUserAnnotations(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "my-ns",
			Annotations: map[string]string{"prometheus.io/scrape": "false",
				"prometheus.io/port":                 "1234",
				"prometheus.io/path":                 "/test",
				"splunk-otel-operator-config/sha256": "shouldBeOverwritten",
			},
		},
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Config: "test",
		}},
	}

	// test
	annotations := Annotations(otelcol)

	//verify
	assert.Equal(t, "false", annotations["prometheus.io/scrape"])
	assert.Equal(t, "1234", annotations["prometheus.io/port"])
	assert.Equal(t, "/test", annotations["prometheus.io/path"])
	assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", annotations["splunk-otel-operator-config/sha256"])
}

func TestAnnotationsPropagateDown(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"myapp": "mycomponent"},
		},
	}

	// test
	annotations := Annotations(otelcol)

	// verify
	assert.Len(t, annotations, 5)
	assert.Equal(t, "mycomponent", annotations["myapp"])
}
