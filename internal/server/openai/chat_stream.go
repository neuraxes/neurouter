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
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

// contentBlockState tracks how deltas for an open content block map to the
// OpenAI chat completion wire format.
type contentBlockState struct {
	isReasoning bool
	toolOrdinal int
}

type chatCompletionStreamServer struct {
	v1.Chat_ChatStreamServer
	ctx     context.Context
	httpCtx http.Context
	buffer  *bytes.Buffer

	id        string
	model     string
	status    v1.ChatStatus
	usage     *v1.Usage
	usageSent bool
	toolCount int
	blocks    map[uint32]contentBlockState
}

func (c *chatCompletionStreamServer) Context() context.Context {
	return c.ctx
}

func (c *chatCompletionStreamServer) accumulateUsage(usage *v1.Usage) {
	if usage == nil {
		return
	}
	if c.usage == nil {
		c.usage = &v1.Usage{}
	}
	if usage.InputTokens != 0 {
		c.usage.InputTokens = usage.InputTokens
	}
	if usage.OutputTokens != 0 {
		c.usage.OutputTokens = usage.OutputTokens
	}
	if usage.CachedInputTokens != 0 {
		c.usage.CachedInputTokens = usage.CachedInputTokens
	}
	if usage.ReasoningTokens != 0 {
		c.usage.ReasoningTokens = usage.ReasoningTokens
	}
}

func (c *chatCompletionStreamServer) writeChunk(chunk *chatCompletionChunk) error {
	chunk.ID = c.id
	chunk.Object = "chat.completion.chunk"
	chunk.Model = c.model
	if chunk.Choices == nil {
		chunk.Choices = []chatCompletionChunkChoice{}
	}

	chunkJson, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	data := append([]byte("data: "), chunkJson...)
	data = append(data, '\n', '\n')

	if c.buffer != nil {
		c.buffer.Write(data)
	}

	_, err = c.httpCtx.Response().Write(data)
	if err != nil {
		return err
	}
	c.httpCtx.Response().(http.Flusher).Flush()
	return nil
}

func (c *chatCompletionStreamServer) writeUsageChunk() error {
	if c.usage == nil || c.usageSent {
		return nil
	}
	c.usageSent = true
	return c.writeChunk(&chatCompletionChunk{
		Choices: []chatCompletionChunkChoice{},
		Usage:   convertUsageToOpenAIChat(c.usage),
	})
}

func (c *chatCompletionStreamServer) getBlockState(index uint32) contentBlockState {
	if c.blocks == nil {
		return contentBlockState{}
	}
	if block, ok := c.blocks[index]; ok {
		return block
	}
	return contentBlockState{}
}

func (c *chatCompletionStreamServer) setBlockState(index uint32, block contentBlockState) {
	if c.blocks == nil {
		c.blocks = map[uint32]contentBlockState{}
	}
	c.blocks[index] = block
}

func (c *chatCompletionStreamServer) Send(event *v1.ChatEvent) error {
	c.accumulateUsage(event.Usage)

	switch e := event.Event.(type) {
	case *v1.ChatEvent_MessageStart:
		c.id = e.MessageStart.GetId()
		c.model = e.MessageStart.GetModel()
		chunk := &chatCompletionChunk{
			Choices: []chatCompletionChunkChoice{{
				Delta: chatCompletionChunkDelta{Role: "assistant"},
			}},
		}
		if err := c.writeChunk(chunk); err != nil {
			return err
		}

	case *v1.ChatEvent_ContentStart:
		start := e.ContentStart
		switch ct := start.Content.(type) {
		case *v1.ContentStart_ToolUse:
			ordinal := c.toolCount
			c.toolCount++
			c.setBlockState(start.GetIndex(), contentBlockState{
				toolOrdinal: ordinal,
			})
			chunk := &chatCompletionChunk{
				Choices: []chatCompletionChunkChoice{{
					Delta: chatCompletionChunkDelta{
						ToolCalls: []toolCall{{
							Index: &ordinal,
							ID:    ct.ToolUse.GetId(),
							Type:  "function",
							Function: functionCall{
								Name: ct.ToolUse.GetName(),
							},
						}},
					},
				}},
			}
			if err := c.writeChunk(chunk); err != nil {
				return err
			}
		default:
			c.setBlockState(start.GetIndex(), contentBlockState{
				isReasoning: start.GetPhase() == v1.ContentPhase_CONTENT_PHASE_REASONING,
			})
		}

	case *v1.ChatEvent_ContentDelta:
		delta := e.ContentDelta
		block := c.getBlockState(delta.GetIndex())
		switch d := delta.Delta.(type) {
		case *v1.ContentDelta_Text:
			chunkDelta := chatCompletionChunkDelta{}
			if block.isReasoning {
				chunkDelta.ReasoningContent = d.Text
			} else {
				chunkDelta.Content = d.Text
			}
			chunk := &chatCompletionChunk{
				Choices: []chatCompletionChunkChoice{{Delta: chunkDelta}},
			}
			if err := c.writeChunk(chunk); err != nil {
				return err
			}
		case *v1.ContentDelta_ToolInputText:
			ordinal := block.toolOrdinal
			chunk := &chatCompletionChunk{
				Choices: []chatCompletionChunkChoice{{
					Delta: chatCompletionChunkDelta{
						ToolCalls: []toolCall{{
							Index:    &ordinal,
							Function: functionCall{Arguments: d.ToolInputText},
						}},
					},
				}},
			}
			if err := c.writeChunk(chunk); err != nil {
				return err
			}
		case *v1.ContentDelta_Signature:
			// OpenAI has no equivalent for reasoning signatures.
		}

	case *v1.ChatEvent_ContentStop:
		if c.blocks != nil {
			delete(c.blocks, e.ContentStop.GetIndex())
		}

	case *v1.ChatEvent_ContentSnapshot:
		// Snapshots (e.g. encrypted reasoning) have no OpenAI representation.

	case *v1.ChatEvent_MessageStop:
		c.status = e.MessageStop.GetStatus()
		chunk := &chatCompletionChunk{
			Choices: []chatCompletionChunkChoice{{
				FinishReason: convertStatusToOpenAIChat(c.status),
			}},
		}
		if err := c.writeChunk(chunk); err != nil {
			return err
		}
		if err := c.writeUsageChunk(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chatCompletionStreamServer) sendDone() error {
	data := []byte("data: [DONE]\n\n")
	if c.buffer != nil {
		c.buffer.Write(data)
	}
	_, err := c.httpCtx.Response().Write(data)
	if err != nil {
		return err
	}
	c.httpCtx.Response().(http.Flusher).Flush()
	return nil
}
