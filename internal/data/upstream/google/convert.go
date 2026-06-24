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

package google

import (
	"encoding/base64"
	"encoding/json"

	"google.golang.org/genai"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
	"github.com/neuraxes/neurouter/internal/util"
)

func convertEffortToGoogleThinkingLevel(effort v1.ReasoningEffort) genai.ThinkingLevel {
	switch effort {
	case v1.ReasoningEffort_REASONING_EFFORT_MINIMAL:
		return genai.ThinkingLevelMinimal
	case v1.ReasoningEffort_REASONING_EFFORT_LOW:
		return genai.ThinkingLevelLow
	case v1.ReasoningEffort_REASONING_EFFORT_MEDIUM:
		return genai.ThinkingLevelMedium
	case v1.ReasoningEffort_REASONING_EFFORT_HIGH, v1.ReasoningEffort_REASONING_EFFORT_EXTRA_HIGH, v1.ReasoningEffort_REASONING_EFFORT_MAX:
		return genai.ThinkingLevelHigh
	default:
		return genai.ThinkingLevelUnspecified
	}
}

func convertGenerationConfigToGoogle(config *v1.GenerationConfig, googleConfig *genai.GenerateContentConfig) {
	if config == nil || googleConfig == nil {
		return
	}
	if config.MaxTokens != nil {
		googleConfig.MaxOutputTokens = int32(*config.MaxTokens)
	}
	if config.Temperature != nil {
		googleConfig.Temperature = config.Temperature
	}
	if config.TopP != nil {
		googleConfig.TopP = config.TopP
	}
	if config.TopK != nil {
		googleConfig.TopK = new(float32(*config.TopK))
	}
	if c := config.ReasoningConfig; c != nil && c.Effort != v1.ReasoningEffort_REASONING_EFFORT_NONE {
		gc := &genai.ThinkingConfig{
			IncludeThoughts: true,
		}
		if c.Effort > v1.ReasoningEffort_REASONING_EFFORT_NONE {
			gc.ThinkingLevel = convertEffortToGoogleThinkingLevel(c.Effort)
		}
		if c.TokenBudget != 0 {
			gc.ThinkingBudget = new(int32(c.TokenBudget))
		}
		googleConfig.ThinkingConfig = gc
	}
	if len(config.StopSequences) > 0 {
		googleConfig.StopSequences = config.StopSequences
	}
}

func convertToolsToGoogle(tools []*v1.Tool) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}
	var functionDecls []*genai.FunctionDeclaration
	for _, tool := range tools {
		switch t := tool.Tool.(type) {
		case *v1.Tool_Function_:
			var parametersJsonSchema any
			if schema := t.Function.GetInputSchema(); schema != nil {
				parametersJsonSchema = schema.AsMap()
			}
			functionDecls = append(functionDecls, &genai.FunctionDeclaration{
				Name:                 t.Function.GetName(),
				Description:          t.Function.GetDescription(),
				ParametersJsonSchema: parametersJsonSchema,
			})
		}
	}
	return []*genai.Tool{{FunctionDeclarations: functionDecls}}
}

func convertContentToGoogle(content *v1.Content) *genai.Part {
	if content.IsReasoning() {
		// Reasoning content should be ignored
		return nil
	}

	switch c := content.Content.(type) {
	case *v1.Content_Text:
		return genai.NewPartFromText(c.Text.GetText())
	case *v1.Content_Image:
		mimeType := c.Image.MimeType
		switch source := c.Image.Source.(type) {
		case *v1.Image_Url:
			return genai.NewPartFromURI(source.Url, mimeType)
		case *v1.Image_Data:
			if mimeType == "" {
				mimeType = util.InferImageMimeType(source.Data)
			}
			return genai.NewPartFromBytes(source.Data, mimeType)
		case *v1.Image_Base64:
			data, err := base64.StdEncoding.DecodeString(source.Base64)
			if err != nil {
				return nil
			}
			if mimeType == "" {
				mimeType = util.InferImageMimeType(data)
			}
			return genai.NewPartFromBytes(data, mimeType)
		default:
			return nil
		}
	case *v1.Content_ToolUse:
		textualInput := c.ToolUse.GetTextualInput()

		var args map[string]any
		if textualInput != "" {
			if err := json.Unmarshal([]byte(textualInput), &args); err != nil {
				args = map[string]any{
					"args": textualInput,
				}
			}
		}

		return genai.NewPartFromFunctionCall(c.ToolUse.Name, args)
	case *v1.Content_ToolResult:
		textualOutput := c.ToolResult.GetTextualOutput()

		var output map[string]any
		if textualOutput != "" {
			if err := json.Unmarshal([]byte(textualOutput), &output); err != nil {
				output = map[string]any{
					"result": textualOutput,
				}
			}
		}

		return genai.NewPartFromFunctionResponse(c.ToolResult.Id, output)
	default:
		return nil
	}
}

func convertMessageToGoogle(msg *v1.Message) *genai.Content {
	var parts []*genai.Part
	for _, content := range msg.Contents {
		if part := convertContentToGoogle(content); part != nil {
			if content.Signature != "" {
				sig, err := base64.StdEncoding.DecodeString(content.Signature)
				if err == nil {
					part.ThoughtSignature = sig
				}
			}
			parts = append(parts, part)
		}
	}

	role := ""
	switch msg.Role {
	case v1.Role_SYSTEM:
		role = "user" // Google AI doesn't support system role, use user role instead
	case v1.Role_USER:
		role = "user"
	case v1.Role_MODEL:
		role = "model"
	}

	return &genai.Content{
		Parts: parts,
		Role:  role,
	}
}

func (r *upstream) convertSystemInstructionToGoogle(messages []*v1.Message) (content *genai.Content) {
	if r.config.SystemAsUser {
		return
	}
	content = &genai.Content{}
	for _, msg := range messages {
		if msg.Role == v1.Role_SYSTEM {
			for _, c := range msg.Contents {
				part := convertContentToGoogle(c)
				if part != nil {
					content.Parts = append(content.Parts, part)
				}
			}
		}
	}
	if len(content.Parts) == 0 {
		return nil
	}
	return
}

// convertStatusFromGoogle maps Google FinishReason to internal ChatStatus.
// It also checks the content to determine if there are pending function calls.
func convertStatusFromGoogle(reason genai.FinishReason, content *genai.Content) v1.ChatStatus {
	// Check if the response contains function calls (tool use)
	if reason == genai.FinishReasonStop && content != nil {
		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				return v1.ChatStatus_CHAT_PENDING_TOOL_USE
			}
		}
	}

	switch reason {
	case genai.FinishReasonStop:
		return v1.ChatStatus_CHAT_COMPLETED
	case genai.FinishReasonMaxTokens:
		return v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT
	case genai.FinishReasonSafety,
		genai.FinishReasonBlocklist,
		genai.FinishReasonProhibitedContent,
		genai.FinishReasonSPII,
		genai.FinishReasonImageSafety,
		genai.FinishReasonImageProhibitedContent:
		return v1.ChatStatus_CHAT_REFUSED
	default:
		return v1.ChatStatus_CHAT_IN_PROGRESS
	}
}

func convertMessageFromGoogle(content *genai.Content) *v1.Message {
	message := &v1.Message{
		Role: v1.Role_MODEL,
	}

	for _, part := range content.Parts {
		var content *v1.Content

		if part.Text != "" {
			phase := v1.ContentPhase_CONTENT_PHASE_NORMAL
			if part.Thought {
				phase = v1.ContentPhase_CONTENT_PHASE_REASONING
			}
			content = &v1.Content{
				Phase:   phase,
				Content: v1.NewTextContent(part.Text),
			}
		} else if part.FunctionCall != nil {
			args, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				continue
			}
			content = &v1.Content{
				Content: &v1.Content_ToolUse{
					ToolUse: &v1.ToolUse{
						Id:   part.FunctionCall.Name,
						Name: part.FunctionCall.Name,
						Inputs: []*v1.ToolUse_Input{
							{
								Input: &v1.ToolUse_Input_Text{
									Text: string(args),
								},
							},
						},
					},
				},
			}
		}

		if content == nil {
			continue
		}

		if len(part.ThoughtSignature) > 0 {
			content.Signature = base64.StdEncoding.EncodeToString(part.ThoughtSignature)
		}

		message.Contents = append(message.Contents, content)
	}

	return message
}

func convertStatisticsFromGoogle(usage *genai.GenerateContentResponseUsageMetadata) *v1.Statistics {
	if usage == nil {
		return nil
	}

	return &v1.Statistics{
		Usage: &v1.Usage{
			InputTokens:       uint32(usage.PromptTokenCount),
			OutputTokens:      uint32(usage.CandidatesTokenCount + usage.ThoughtsTokenCount),
			CachedInputTokens: uint32(usage.CachedContentTokenCount),
			ReasoningTokens:   uint32(usage.ThoughtsTokenCount),
		},
	}
}
