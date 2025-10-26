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

package repository

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/neuraxes/neurouter/internal/conf"
)

// UpstreamConfig is a type constraint for LLM provider configurations.
// It allows for configuration of different upstream LLM providers like OpenAI, Google, Anthropic, etc.
type UpstreamConfig interface {
	conf.NeurouterConfig | conf.OpenAIConfig | conf.GoogleConfig | conf.AnthropicConfig
}

// UpstreamFactory is a generic factory function type for creating ChatRepo instances.
type UpstreamFactory[T UpstreamConfig] func(config *T, logger log.Logger) (ChatRepo, error)
