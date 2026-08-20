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
	"net"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/semconv/v1.41.0/genaiconv"

	"github.com/neuraxes/neurouter/internal/biz/observability"
	"github.com/neuraxes/neurouter/internal/conf"
)

func (m *model) GenAITarget() observability.GenAITarget {
	upstreamModel := m.config.GetUpstreamId()
	if upstreamModel == "" {
		upstreamModel = m.config.GetId()
	}

	var address string
	var port int
	switch config := m.upstreamConfig.GetConfig().(type) {
	case *conf.UpstreamConfig_OpenAi:
		address, port = parseServer(config.OpenAi.GetBaseUrl())
	case *conf.UpstreamConfig_Anthropic:
		address, port = parseServer(config.Anthropic.GetBaseUrl())
	case *conf.UpstreamConfig_Neurouter:
		address, port = parseServer(config.Neurouter.GetEndpoint())
	}

	return observability.GenAITarget{
		Provider:      genaiconv.ProviderNameAttr(m.config.GetProvider()),
		Upstream:      m.upstreamConfig.GetName(),
		RouterModel:   m.config.GetId(),
		UpstreamModel: upstreamModel,
		ServerAddress: address,
		ServerPort:    port,
	}
}

// parseServer splits an endpoint into host and port, accepting both URLs and the
// host:port and dns:///host:port forms used by gRPC endpoints. The port is zero
// when the endpoint omits it; no scheme default is assumed.
func parseServer(endpoint string) (string, int) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return "", 0
	}

	if parsed, err := url.Parse(endpoint); err == nil && parsed.Hostname() != "" {
		port, _ := strconv.Atoi(parsed.Port())
		return parsed.Hostname(), port
	}

	endpoint = strings.TrimPrefix(endpoint, "dns:///")
	if host, portText, err := net.SplitHostPort(endpoint); err == nil {
		port, _ := strconv.Atoi(portText)
		return host, port
	}

	return strings.Trim(endpoint, "[]"), 0
}
