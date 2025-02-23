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

package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
	. "github.com/signalfx/splunk-otel-collector-operator/internal/collector"
)

var logger = logf.Log.WithName("unit-tests")

func TestContainerNewDefault(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Equal(t, "quay.io/signalfx/splunk-otel-collector:0.0.0", c.Image)
}

func TestContainerConfigFlagIsIgnored(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Args: map[string]string{
				"key":    "value",
				"config": "/some-custom-file.yaml",
			},
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Len(t, c.Args, 2)
	assert.Contains(t, c.Args, "--key=value")
	assert.NotContains(t, c.Args, "--config=/some-custom-file.yaml")
}

func TestContainerCustomVolumes(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			VolumeMounts: []corev1.VolumeMount{{
				Name: "custom-volume-mount",
			}},
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Len(t, c.VolumeMounts, 2)
	assert.Equal(t, "custom-volume-mount", c.VolumeMounts[1].Name)
}

func TestContainerCustomSecurityContext(t *testing.T) {
	// default config without security context
	c1 := Container(logger, v1alpha1.SplunkCollectorSpec{})

	// verify
	assert.Nil(t, c1.SecurityContext)

	// prepare
	isPrivileged := true
	uid := int64(1234)

	// test
	c2 := Container(logger, v1alpha1.SplunkCollectorSpec{
		SecurityContext: &corev1.SecurityContext{
			Privileged: &isPrivileged,
			RunAsUser:  &uid,
		},
	})

	// verify
	assert.NotNil(t, c2.SecurityContext)
	assert.True(t, *c2.SecurityContext.Privileged)
	assert.Equal(t, *c2.SecurityContext.RunAsUser, uid)
}

func TestContainerEnvVarsOverridden(t *testing.T) {
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Env: []corev1.EnvVar{
				{
					Name:  "foo",
					Value: "bar",
				},
			},
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Len(t, c.Env, 1)
	assert.Equal(t, "foo", c.Env[0].Name)
	assert.Equal(t, "bar", c.Env[0].Value)
}

func TestContainerEmptyEnvVarsByDefault(t *testing.T) {
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Empty(t, c.Env)
}

func TestContainerResourceRequirements(t *testing.T) {
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128M"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256M"),
				},
			},
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Equal(t, resource.MustParse("100m"), *c.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("128M"), *c.Resources.Limits.Memory())
	assert.Equal(t, resource.MustParse("200m"), *c.Resources.Requests.Cpu())
	assert.Equal(t, resource.MustParse("256M"), *c.Resources.Requests.Memory())
}

func TestContainerDefaultResourceRequirements(t *testing.T) {
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Empty(t, c.Resources)
}

func TestContainerArgs(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Args: map[string]string{
				"metrics-level": "detailed",
				"log-level":     "debug",
			},
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Contains(t, c.Args, "--metrics-level=detailed")
	assert.Contains(t, c.Args, "--log-level=debug")
}

func TestContainerImagePullPolicy(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			ImagePullPolicy: corev1.PullIfNotPresent,
		}},
	}

	// test
	c := Container(logger, otelcol.Spec.Agent)

	// verify
	assert.Equal(t, c.ImagePullPolicy, corev1.PullIfNotPresent)
}
