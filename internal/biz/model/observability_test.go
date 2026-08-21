// Copyright 2024 Neurouter Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"

	"github.com/neuraxes/neurouter/internal/conf"
)

func TestModelGenAITarget(t *testing.T) {
	Convey("Given an OpenAI-compatible model with an Azure provider", t, func() {
		m := &model{
			config: &conf.Model{
				Id:         "router-gpt",
				UpstreamId: "gpt-4.1",
				Provider:   "azure",
			},
			upstreamConfig: &conf.UpstreamConfig{
				Name: "azure-primary",
				Config: &conf.UpstreamConfig_OpenAi{OpenAi: &conf.OpenAIConfig{
					BaseUrl: "https://example.openai.azure.com:8443/openai/v1",
				}},
			},
		}

		target := m.GenAITarget()

		So(target.Provider, ShouldEqual, genaiconv.ProviderNameAzureAIOpenAI)
		So(target.Upstream, ShouldEqual, "azure-primary")
		So(target.Model, ShouldEqual, "router-gpt")
		So(target.UpstreamModel, ShouldEqual, "gpt-4.1")
		So(target.ServerAddress, ShouldEqual, "example.openai.azure.com")
		So(target.ServerPort, ShouldEqual, 8443)
	})

	Convey("Given a chained Neurouter model without an explicit provider", t, func() {
		m := &model{
			config: &conf.Model{Id: "remote-model"},
			upstreamConfig: &conf.UpstreamConfig{
				Name: "remote",
				Config: &conf.UpstreamConfig_Neurouter{Neurouter: &conf.NeurouterConfig{
					Endpoint: "dns:///router.internal:9000",
				}},
			},
		}

		target := m.GenAITarget()

		So(target.Provider, ShouldEqual, genaiconv.ProviderNameAttr("neurouter"))
		So(target.UpstreamModel, ShouldEqual, "remote-model")
		So(target.ServerAddress, ShouldEqual, "router.internal")
		So(target.ServerPort, ShouldEqual, 9000)
	})
}
