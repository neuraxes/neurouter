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

package v1

type ChatEventPayload interface {
	isChatEvent_Event
}

func NewChatEvent(id string, event ChatEventPayload) *ChatEvent {
	return &ChatEvent{Id: id, Event: event}
}

// NewMessageStartEvent opens a model turn carrying the message identity.
func NewMessageStartEvent(messageID, model string) *ChatEvent_MessageStart {
	return &ChatEvent_MessageStart{
		MessageStart: &MessageStart{Id: messageID, Model: model},
	}
}

// NewMessageStopEvent terminates a model turn with the final status.
func NewMessageStopEvent(status ChatStatus) *ChatEvent_MessageStop {
	return &ChatEvent_MessageStop{
		MessageStop: &MessageStop{Status: status},
	}
}

// NewContentStartTextEvent opens a streamable text content block.
func NewContentStartTextEvent(index uint32, phase ContentPhase) *ChatEvent_ContentStart {
	return &ChatEvent_ContentStart{
		ContentStart: &ContentStart{
			Index:   index,
			Phase:   phase,
			Content: &ContentStart_Text{Text: &TextStart{}},
		},
	}
}

// NewContentStartToolUseEvent opens a streamable tool-use content block.
// The content phase is NORMAL by default.
func NewContentStartToolUseEvent(index uint32, id, name string) *ChatEvent_ContentStart {
	return &ChatEvent_ContentStart{
		ContentStart: &ContentStart{
			Index:   index,
			Content: &ContentStart_ToolUse{ToolUse: &ToolUseStart{Id: id, Name: name}},
		},
	}
}

// NewContentDeltaTextEvent carries a text or reasoning fragment.
func NewContentDeltaTextEvent(index uint32, text string) *ChatEvent_ContentDelta {
	return &ChatEvent_ContentDelta{
		ContentDelta: &ContentDelta{Index: index, Delta: &ContentDelta_Text{Text: text}},
	}
}

// NewContentDeltaSignatureEvent carries a verification-signature fragment.
func NewContentDeltaSignatureEvent(index uint32, signature string) *ChatEvent_ContentDelta {
	return &ChatEvent_ContentDelta{
		ContentDelta: &ContentDelta{Index: index, Delta: &ContentDelta_Signature{Signature: signature}},
	}
}

// NewContentDeltaToolInputTextEvent carries a tool-use input text fragment.
func NewContentDeltaToolInputTextEvent(index uint32, text string) *ChatEvent_ContentDelta {
	return &ChatEvent_ContentDelta{
		ContentDelta: &ContentDelta{Index: index, Delta: &ContentDelta_ToolInputText{ToolInputText: text}},
	}
}

// NewContentStopEvent closes the content block at the given index.
func NewContentStopEvent(index uint32) *ChatEvent_ContentStop {
	return &ChatEvent_ContentStop{ContentStop: &ContentStop{Index: index}}
}

// NewContentSnapshotEvent carries a full content, normally for non-streamable content.
func NewContentSnapshotEvent(content *Content) *ChatEvent_ContentSnapshot {
	return &ChatEvent_ContentSnapshot{ContentSnapshot: content}
}
