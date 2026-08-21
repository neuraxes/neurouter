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
	address, port := upstreamServer(m.upstreamConfig)
	return observability.GenAITarget{
		Provider:      genAIProvider(m.config.GetProvider(), m.upstreamConfig),
		Upstream:      m.upstreamConfig.GetName(),
		Model:         m.config.GetId(),
		UpstreamModel: upstreamModel,
		ServerAddress: address,
		ServerPort:    port,
	}
}

func genAIProvider(provider string, upstream *conf.UpstreamConfig) genaiconv.ProviderNameAttr {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "azure", "azure-openai", "azure_openai", "azure.ai.openai":
		return genaiconv.ProviderNameAzureAIOpenAI
	case "google", "gemini", "gcp.gemini":
		return genaiconv.ProviderNameGCPGemini
	case "xai", "x.ai", "x_ai":
		return genaiconv.ProviderNameXAI
	case "mistral", "mistral-ai", "mistral_ai":
		return genaiconv.ProviderNameMistralAI
	case "openai":
		return genaiconv.ProviderNameOpenAI
	case "anthropic":
		return genaiconv.ProviderNameAnthropic
	case "deepseek":
		return genaiconv.ProviderNameDeepseek
	case "groq":
		return genaiconv.ProviderNameGroq
	case "perplexity":
		return genaiconv.ProviderNamePerplexity
	case "":
		switch upstream.GetConfig().(type) {
		case *conf.UpstreamConfig_OpenAi:
			return genaiconv.ProviderNameOpenAI
		case *conf.UpstreamConfig_Google:
			return genaiconv.ProviderNameGCPGemini
		case *conf.UpstreamConfig_Anthropic:
			return genaiconv.ProviderNameAnthropic
		case *conf.UpstreamConfig_Neurouter:
			return genaiconv.ProviderNameAttr("neurouter")
		default:
			return genaiconv.ProviderNameAttr("unknown")
		}
	default:
		return genaiconv.ProviderNameAttr(provider)
	}
}

func upstreamServer(upstream *conf.UpstreamConfig) (string, int) {
	var endpoint string
	switch config := upstream.GetConfig().(type) {
	case *conf.UpstreamConfig_OpenAi:
		endpoint = config.OpenAi.GetBaseUrl()
		if endpoint == "" {
			endpoint = "https://api.openai.com"
		}
	case *conf.UpstreamConfig_Google:
		endpoint = "https://generativelanguage.googleapis.com"
	case *conf.UpstreamConfig_Anthropic:
		endpoint = config.Anthropic.GetBaseUrl()
		if endpoint == "" {
			endpoint = "https://api.anthropic.com"
		}
	case *conf.UpstreamConfig_Neurouter:
		endpoint = config.Neurouter.GetEndpoint()
	}
	return parseServer(endpoint)
}

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
