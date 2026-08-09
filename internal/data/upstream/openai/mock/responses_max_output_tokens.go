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

//go:embed responses_max_output_tokens_request.json
var responsesMaxOutputTokensRequest []byte

//go:embed responses_max_output_tokens_response.json
var responsesMaxOutputTokensResponse []byte

// ResponsesMaxOutputTokens covers an incomplete response caused by the output
// token limit.
var ResponsesMaxOutputTokens = &Fixture{
	Name:     "responses_max_output_tokens",
	Request:  responsesMaxOutputTokensRequest,
	Response: responsesMaxOutputTokensResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_max_output_tokens",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(64)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_MINIMAL},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Answer in long, detailed prose.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Write a detailed 300-word explanation of how an LLM router balances load across multiple upstream providers, covering probing, ranking, reservation, and rate limiting.")},
				},
			},
		},
	},
	ChatResp: &v1.ChatResp{
		Id:     "responses_max_output_tokens",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_REACHED_TOKEN_LIMIT,
		Message: &v1.Message{
			Id:   "gen-1785502363-1OIGKSAZ6DupYJGniSP8",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_0ceb79d4157e81e4016a6c9a9c0b9c81a1aaba69cfa1664a3d",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJqcy2ECr5Rg0-Fawu_sXUVrhTJh5tGMvz7ilDu1PtqkBiDRq4pmkQBCtA-yTpBV-fmKs7r6cD0TvrBC-7FnOxkDSrYRcdqwBQV0ZJP5V5Y0E_PdtHPkFPahDvYePNgl5R-SeQnieBaWhDdIQOqM92duvquQNcsR_ccUliY61_97179tnx2GleNy1l_2_NbLUQcYWSzJR8BMckvqCNcmoo5V-eIZF3WX6VItlwkUhWOTDsPGSyRsJlKbiuZLAHo-zxEwKX0slRcE7N0bXSdnTFsd7Ir7EqLDIAznzWoka6OI--O_LZgUDEsl8DduwTSwATsoMZ4MX-DfNha1fg24gC1Ll6u_Exmw1fUSlfWFY9MLyjbsLJATnTGh3eco7yKQCXedaSdYCpJ-PAVZKkmFml9cV9bVxRkZl5qDHdayAx-nu8kSy4TLdhM7vMQS6hHF9ptzBI66DoHfSIZRBkuabzqJs6PsiF_aZ1PE9I-UzM9i4FkOufwAztFa0a38Q7CmjkyLgnahIXUSJC0bQPqE3NgCFZvvz_ETUYDh9JDr6msWotltUKl1l_yTP9zj8yabKoyRCMr89ehK82RQWgWiV4Lx0ul2khBEjCwRKNfaYi923Bt_PLAZH_-JIawtzapPxNcyTXRuOeI7gDC18NEfGfprSNEecOaJoCch5TQ819dGh2EuzOE1h_OXgojmjvMNtHL_pJedo366u-Squ9ouTnF44GP7b7-luCNYe2gdpBgks2S3te8rUKQngA4AJ_IyDEDJ8YSAugyWW6ZbZLG_NzJdEY3J165v0zal98-ilfd1i8rR3xtc1CKqa3p6yB0C5uHN8IoQUYqyGSAHeBLxGu2AsBvyulaNXWx7pJS9GDSw-GW3XnB5xNR5tud1hrZBUCmZ3ol9W8PnZ4trR3RZxib7K0C9jjGpvWXRX85fk7uSuybBB9MpHPjou808EhurDl4QNDAAHywIKuZPiGysDA4aJiFQrNJ9x3js2RG_RlLiJuY=",
					},
				},
				{Content: v1.NewTextContent("An LLM router balances load across multiple upstream providers by implementing a coordinated sequence of probing, ranking, reservation, and rate-limiting mechanisms that together maximize throughput, minimize latency, and ensure reliability and cost-effectiveness. First, continuous")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 55, OutputTokens: 64},
		},
	},
}
