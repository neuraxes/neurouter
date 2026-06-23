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

package openai

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"google.golang.org/protobuf/types/known/structpb"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

// convertEffortToOpenAI maps the internal reasoning effort to the OpenAI shared reasoning effort.
func convertEffortToOpenAI(effort v1.ReasoningEffort) shared.ReasoningEffort {
	switch effort {
	case v1.ReasoningEffort_REASONING_EFFORT_NONE:
		return shared.ReasoningEffortNone
	case v1.ReasoningEffort_REASONING_EFFORT_MINIMAL:
		return shared.ReasoningEffortMinimal
	case v1.ReasoningEffort_REASONING_EFFORT_LOW:
		return shared.ReasoningEffortLow
	case v1.ReasoningEffort_REASONING_EFFORT_MEDIUM:
		return shared.ReasoningEffortMedium
	case v1.ReasoningEffort_REASONING_EFFORT_HIGH:
		return shared.ReasoningEffortHigh
	case v1.ReasoningEffort_REASONING_EFFORT_EXTRA_HIGH, v1.ReasoningEffort_REASONING_EFFORT_MAX:
		return shared.ReasoningEffortXhigh
	default:
		return ""
	}
}

// convertImageToOpenAIURL converts an internal image to the URL form OpenAI expects,
// inlining raw or base64 data as a data URL.
func convertImageToOpenAIURL(image *v1.Image) string {
	if image == nil {
		return ""
	}
	switch image.Source.(type) {
	case *v1.Image_Url:
		return image.GetUrl()
	case *v1.Image_Data:
		return util.EncodeImageDataToURL(image.GetData(), image.GetMimeType())
	case *v1.Image_Base64:
		return util.EncodeImageBase64ToURL(image.GetBase64(), image.GetMimeType())
	}
	return ""
}

func convertSchemaToMap(schema *structpb.Struct) openai.FunctionParameters {
	if schema == nil {
		return nil
	}
	return openai.FunctionParameters(schema.AsMap())
}
