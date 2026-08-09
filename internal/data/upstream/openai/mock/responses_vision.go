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

package mock

import (
	_ "embed"

	v1 "github.com/neuraxes/neurouter/api/neurouter/v1"
)

//go:embed responses_vision_request.json
var responsesVisionRequest []byte

//go:embed responses_vision_response.json
var responsesVisionResponse []byte

// ResponsesVision covers inline base64 image input.
var ResponsesVision = &Fixture{
	Name:     "responses_vision",
	Request:  responsesVisionRequest,
	Response: responsesVisionResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_vision",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(1024)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_MINIMAL},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Inspect every image the user supplies and describe it briefly.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("The image below is provided as inline base64 data. State its source (base64) and give a one-sentence description of what you see, including any visible logo or shape.")},
					{
						Content: &v1.Content_Image{
							Image: &v1.Image{
								MimeType: "image/png",
								Source:   &v1.Image_Base64{Base64: openAILogoPNG},
							},
						},
					},
				},
			},
		},
	},
	ChatResp: &v1.ChatResp{
		Id:     "responses_vision",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1785502390-uWUQAQNqsj1Ocw0q6uQk",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_01969eef50f6c980016a6c9ab75de8819fa975bc775b833f4f",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJq3gNByFtjiMDsRgZymiPXje54G2e6_ufgtt1Mtm-Jqoky1_nFAkU3kVvqd457ZZHaoyJx3aMrz0k9XafMWdLId08khC_EwSg2zM0YBZd-hg1MMtjSCsN__LZpeS4gUnKcLtnc5hDwSOLqwqpCPf5NK1X3TwMVR1gNAcki-0CLvKvlPZhazC5zO9wXMb7GPVo4fowRTW25JZeCllBB86F3omukVqxqNMMg-8Dx8j1IXvwjC51kXCKNAPbqBPj0ogiRpQZUf95wnxC4PEWdgS18o6KsHfxMtbFPGkVlwth0UWnEt9D_Mo0h6NFClgUfmqqYJmENA4sk-epoP7ma6TMQxPHalr0T4ScmZtdfTBt-tkjUG6YOeiTtP5DXXdGhpqNDalhGon20Gs2zP33Ejxlyw4EhyapmavWpX6BS2HhuFvQMxurIr_AL1PHhRSt7o1tWCYOEy7mFgz6LvaZxGr95BgRd-DrQh_QYmery7BeEzrxS4dTwqeTsFSAnJsUGaipElJVYenR35FP9CCjNrfR65o-Uc0RQGgWydarM4QxcX0Kf4NV_kzesJTOhqUQRrtYi_pr9iOMWJeYonbmXAyzkxcLZmvH1_NcBW5sxIafz0x5WZ7oEfJoJXN7fiGAgrp_Dp7NL0L0LInUTwNdSbzjbDomIci17ZNr4OCGLfUQ5o5-k00zSOHtf4egmSNcWyKHySk_gSwr79y9SiptEQbFISyvywyhhY4DnP9NQqCQwjKJiDtN1NEnCRQCWjt0V8-89G8RJqiFXBl7OHsTJXzD5LiPNYTmmSW4sw4HFhtUx9pqirv-zv99KjjYL2VgiLTxkXaEV0yegXmrqzNlFBSUo3UpRer9_9oaLJK8dGgcaBisfWGpIjyqY5DMMMudIvtjybWoxqSv8N6JoP1TVaMhkkhrs-1eL_1-Ytg64UCE8zvGwMQj18hH4-3cASg9rCGBRK4WAPe1Shm5MUOEovXoTGyHS1srfu4BytWE9Wkkfgv_4=",
					},
				},
				{
					Id: "msg_tmp_9wt3kgyj5l",
					Content: v1.NewTextContent(`Source (base64): data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAAnFBMVEUAAAD///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////8/ZvTFAAAAMXRSTlMAAQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyAhIiMlJicoKSorLC0uLzAxMjM0NwAAAEpJREFUGNNjYGBgYGRkYGBgZmBjYGBgYGBgYGJgYGBkYWBkZGBg6GBkYGBgYGBgYGD4DgkAG4YCgABBgA5FQyS9wAAAABJRU5ErkJggg==

One-sentence description: A small square icon showing the OpenAI swirl logo in dark lines on a white background with a thin rounded gray border.`),
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 67, OutputTokens: 223},
		},
	},
}
