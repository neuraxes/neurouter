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

package deepseek

type ChatRequest struct {
	// A list of messages comprising the conversation so far
	Messages []*Message `json:"messages" validate:"required"`
	// ID of the model to use. You can use deepseek-chat
	Model string `json:"model" validate:"required"`
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so
	// far, decreasing the model's likelihood to repeat the same line verbatim
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty" validate:"gte=-2,lte=2"`
	// Integer between 1 and 8192. The maximum number of tokens that can be generated in the chat completion.
	// The total length of input tokens and generated tokens is limited by the model's context length.
	// If max_tokens is not specified, the default value 4096 is used
	MaxTokens int `json:"max_tokens,omitempty" validate:"gte=1,lte=8192"`
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics
	PresencePenalty float64 `json:"presence_penalty,omitempty" validate:"gte=-2,lte=2"`
	// An object specifying the format that the model must output. Setting to {"type": "json_object"} enables JSON Output,
	// which guarantees the message the model generates is valid JSON
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	// Up to 16 sequences where the API will stop generating further tokens
	Stop any `json:"stop,omitempty"`
	// If set, partial message deltas will be sent. Tokens will be sent as data-only server-sent events (SSE) as they
	// become available, with the stream terminated by a data: [DONE] message
	Stream bool `json:"stream,omitempty"`
	// Options for streaming response. Only set this when you set stream: true
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random,
	// while lower values like 0.2 will make it more focused and deterministic
	Temperature float64 `json:"temperature,omitempty" validate:"gte=0,lte=2"`
	// An alternative to sampling with temperature, called nucleus sampling, where the model considers the results
	// of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass
	// are considered
	TopP float64 `json:"top_p,omitempty" validate:"gte=0,lte=1"`
	// A list of tools the model may call. Currently, only functions are supported as a tool. Use this to provide a list
	// of functions the model may generate JSON inputs for. A max of 128 functions are supported.
	Tools []*Tool `json:"tools,omitempty"`
	// Controls which (if any) tool is called by the model
	ToolChoice *ToolChoice `json:"tool_choice,omitempty"`
	// Whether to return log probabilities of the output tokens or not. If true, returns the log probabilities of each
	// output token returned in the content of message
	Logprobs bool `json:"logprobs,omitempty"`
	// An integer between 0 and 20 specifying the number of most likely tokens to return at each token position, each
	// with an associated log probability. logprobs must be set to true if this parameter is used.
	TopLogprobs int `json:"top_logprobs,omitempty" validate:"gte=0,lte=20"`
}

type ResponseFormat struct {
	// Must be one of "text" or "json_object" (default to "text")
	Type string `json:"type" validate:"required"`
}

type StreamOptions struct {
	// If set, an additional chunk will be streamed with usage statistics before the [DONE] message. The usage field
	// shows the token usage statistics for the entire request, and the choices field will be an empty array
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type Tool struct {
	// Currently, only "function" is supported as a tool type
	Type string `json:"type" validate:"required"`
	// The function definition including name, description and parameters schema
	Function *FunctionDefinition `json:"function" validate:"required"`
}

type ToolChoice struct {
	// Must be "function". Currently, only function is supported as a tool
	Type string `json:"type" validate:"required"`
	// Reference to which function to call by name
	Function *FunctionChoice `json:"function,omitempty"`
}

type FunctionDefinition struct {
	// The name of the function to be called. Must be a-z, A-Z, 0-9, or contain underscores and dashes, with a maximum
	// length of 64
	Name string `json:"name" validate:"required"`
	// A description of what the function does, used by the model to choose when and how to call the function
	Description string `json:"description,omitempty"`
	// The parameters the function accepts, described as a JSON Schema object
	Parameters map[string]any `json:"parameters,omitempty"`
}

type FunctionChoice struct {
	// The name of the function to call
	Name string `json:"name" validate:"required"`
}

type ChatResponse struct {
	// A unique identifier for the chat completion
	ID string `json:"id"`
	// A list of chat completion choices. Can include multiple choices if n>1
	Choices []*ChatChoice `json:"choices"`
	// The Unix timestamp (in seconds) of when the chat completion was created
	Created int64 `json:"created"`
	// The model used for the chat completion
	Model string `json:"model"`
	// This fingerprint represents the backend configuration that the model runs with
	SystemFingerprint string `json:"system_fingerprint"`
	// The object type, which is always "chat.completion"
	Object string `json:"object"`
	// Usage statistics for the completion request
	Usage *Usage `json:"usage,omitempty"`
}

type ChatChoice struct {
	// The reason why the model stopped generating tokens
	FinishReason string `json:"finish_reason"`
	// The index of the choice in the list of choices
	Index int `json:"index"`
	// The chat completion message generated by the model
	Message *Message `json:"message"`
	// Log probability information for the choice
	LogProbs *LogProbs `json:"logprobs,omitempty"`
}

type ChatStreamResponse struct {
	// A unique identifier for the chat completion. Each chunk has the same ID
	ID string `json:"id"`
	// A list of chat completion choices for this chunk
	Choices []*ChatStreamChoice `json:"choices"`
	// The Unix timestamp of when the chat completion was created. Each chunk has the same timestamp
	Created int64 `json:"created"`
	// The model used to generate the completion
	Model string `json:"model"`
	// This fingerprint represents the backend configuration that the model runs with
	SystemFingerprint string `json:"system_fingerprint"`
	// The object type, which is always "chat.completion.chunk"
	Object string `json:"object"`
	// Usage statistics included in the final chunk if stream_options.include_usage is true
	Usage *Usage `json:"usage,omitempty"`
}

type ChatStreamChoice struct {
	// Delta message content for this chunk
	Delta *Message `json:"delta"`
	// The reason why the model stopped generating (only in final chunk)
	FinishReason string `json:"finish_reason,omitempty"`
	// The index of the choice in the list of choices
	Index int `json:"index"`
}

type LogProbs struct {
	// A list of message content tokens with log probability information
	Content []TokenLogProb `json:"content"`
}

type TokenLogProb struct {
	// The token
	Token string `json:"token"`
	// The log probability of this token
	LogProb float64 `json:"logprob"`
	// A list of integers representing the UTF-8 bytes representation of the token
	Bytes []int `json:"bytes"`
	// List of the most likely tokens and their log probability at this token position
	TopLogProbs []TokenLogProb `json:"top_logprobs"`
}

type Usage struct {
	// Number of tokens in the generated completion
	CompletionTokens int `json:"completion_tokens"`
	// Number of tokens in the prompt. Equals prompt_cache_hit_tokens + prompt_cache_miss_tokens
	PromptTokens int `json:"prompt_tokens"`
	// Number of tokens in the prompt that hits the context cache
	PromptCacheHitTokens int `json:"prompt_cache_hit_tokens"`
	// Number of tokens in the prompt that misses the context cache
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
	// Total number of tokens used in the request (prompt + completion)
	TotalTokens int `json:"total_tokens"`
	// Breakdown of tokens used in a completion
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

type CompletionTokensDetails struct {
	// Tokens generated by the model for reasoning
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Message struct {
	// The role of the message's author. Must be one of: "system", "user", "assistant", "tool"
	Role string `json:"role" validate:"required"`
	// The contents of the message
	Content string `json:"content" validate:"required"`
	// An optional name for the participant. Provides the model information to differentiate between participants of the
	// same role
	Name string `json:"name,omitempty"`
	// Set this to true to force the model to start its answer by the content of the supplied prefix in this assistant
	// message
	Prefix bool `json:"prefix,omitempty"`
	// Used for the deepseek-reasoner model in the Chat Prefix Completion feature as the input for the CoT in the last
	// assistant message
	ReasoningContent string `json:"reasoning_content,omitempty"`
	// Tool call that this message is responding to
	ToolCallID string `json:"tool_call_id,omitempty"`
	// The tool calls generated by the model, such as function calls
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	// The ID of the tool call
	ID string `json:"id" validate:"required"`
	// The type of the tool. Currently, only "function" is supported
	Type string `json:"type" validate:"required"`
	// The function that the model called
	Function *FunctionCall `json:"function" validate:"required"`
}

type FunctionCall struct {
	// The name of the function to call
	Name string `json:"name" validate:"required"`
	// The arguments to call the function with, as generated by the model in JSON format
	Arguments string `json:"arguments" validate:"required"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (e *Error) Error() string {
	return e.Message
}
