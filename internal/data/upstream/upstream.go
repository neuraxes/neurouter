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

package upstream

import (
	"github.com/google/wire"

	"github.com/neuraxes/neurouter/internal/data/upstream/anthropic"
	"github.com/neuraxes/neurouter/internal/data/upstream/deepseek"
	"github.com/neuraxes/neurouter/internal/data/upstream/google"
	"github.com/neuraxes/neurouter/internal/data/upstream/neurouter"
	"github.com/neuraxes/neurouter/internal/data/upstream/openai"
)

var ProviderSet = wire.NewSet(
	anthropic.NewAnthropicChatRepoFactory,
	deepseek.NewDeepSeekChatRepoFactory,
	google.NewGoogleFactory,
	neurouter.NewNeurouterFactory,
	openai.NewOpenAIChatRepoFactory,
)
