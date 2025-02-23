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

package parser_test

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/signalfx/splunk-otel-collector-operator/internal/collector/parser"
)

var logger = logf.Log.WithName("unit-tests")

func TestParseEndpoint(t *testing.T) {
	// prepare
	// there's no parser registered to handle "myreceiver", so, it falls back to the generic parser
	builder := parser.NewGenericReceiverParser(logger, "myreceiver", map[interface{}]interface{}{
		"endpoint": "0.0.0.0:1234",
	})

	// test
	ports, err := builder.Ports()

	// verify
	assert.NoError(t, err)
	assert.Len(t, ports, 1)
	assert.EqualValues(t, 1234, ports[0].Port)
}

func TestFailedToParseEndpoint(t *testing.T) {
	// prepare
	// there's no parser registered to handle "myreceiver", so, it falls back to the generic parser
	builder := parser.NewGenericReceiverParser(logger, "myreceiver", map[interface{}]interface{}{
		"endpoint": "0.0.0.0",
	})

	// test
	ports, err := builder.Ports()

	// verify
	assert.NoError(t, err)
	assert.Len(t, ports, 0)
}

func TestDownstreamParsers(t *testing.T) {
	for _, tt := range []struct {
		desc         string
		receiverName string
		parserName   string
		defaultPort  int
		builder      func(logr.Logger, string, map[interface{}]interface{}) parser.ReceiverParser
	}{
		{"zipkin", "zipkin", "__zipkin", 9411, parser.NewZipkinReceiverParser},
		{"opencensus", "opencensus", "__opencensus", 55678, parser.NewOpenCensusReceiverParser},

		// contrib receivers
		{"carbon", "carbon", "__carbon", 2003, parser.NewCarbonReceiverParser},
		{"collectd", "collectd", "__collectd", 8081, parser.NewCollectdReceiverParser},
		{"sapm", "sapm", "__sapm", 7276, parser.NewSAPMReceiverParser},
		{"signalfx", "signalfx", "__signalfx", 9943, parser.NewSignalFxReceiverParser},
		{"wavefront", "wavefront", "__wavefront", 2003, parser.NewWavefrontReceiverParser},
		{"zipkin-scribe", "zipkin-scribe", "__zipkinscribe", 9410, parser.NewZipkinScribeReceiverParser},
		{"fluentforward", "fluentforward", "__fluentforward", 8006, parser.NewFluentForwardReceiverParser},
		{"statsd", "statsd", "__statsd", 8125, parser.NewStatsdReceiverParser},
	} {
		t.Run(tt.receiverName, func(t *testing.T) {
			t.Run("builds successfully", func(t *testing.T) {
				// test
				builder := tt.builder(logger, tt.receiverName, map[interface{}]interface{}{})

				// verify
				assert.Equal(t, tt.parserName, builder.ParserName())
			})

			t.Run("assigns the expected port", func(t *testing.T) {
				// prepare
				builder := tt.builder(logger, tt.receiverName, map[interface{}]interface{}{})

				// test
				ports, err := builder.Ports()

				// verify
				assert.NoError(t, err)
				assert.Len(t, ports, 1)
				assert.EqualValues(t, tt.defaultPort, ports[0].Port)
				assert.Equal(t, tt.receiverName, ports[0].Name)
			})

			t.Run("allows port to be overridden", func(t *testing.T) {
				// prepare
				builder := tt.builder(logger, tt.receiverName, map[interface{}]interface{}{
					"endpoint": "0.0.0.0:65535",
				})

				// test
				ports, err := builder.Ports()

				// verify
				assert.NoError(t, err)
				assert.Len(t, ports, 1)
				assert.EqualValues(t, 65535, ports[0].Port)
				assert.Equal(t, tt.receiverName, ports[0].Name)
			})
		})
	}
}
