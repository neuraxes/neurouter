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

//go:embed responses_text_request.json
var responsesTextRequest []byte

//go:embed responses_text_response.json
var responsesTextResponse []byte

// ResponsesText covers a basic Responses API text generation with request
// metadata.
var ResponsesText = &Fixture{
	Name:     "responses_text",
	Request:  responsesTextRequest,
	Response: responsesTextResponse,
	ChatReq: &v1.ChatReq{
		Id:    "responses_text",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(512)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_MINIMAL},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Follow formatting instructions exactly.")},
				},
			},
			{
				Role: v1.Role_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("Reply with exactly one sentence: Neurouter successfully routed this Responses API request.")},
				},
			},
		},
		Metadata: map[string]string{"fixture": "responses_text"},
	},
	ChatResp: &v1.ChatResp{
		Id:     "gen-1785502322-iD3pWMHAKB2iRN8OtooU",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_COMPLETED,
		Message: &v1.Message{
			// A Responses output is a heterogeneous item list, so the response ID
			// identifies the combined internal model message.
			Id:   "gen-1785502322-iD3pWMHAKB2iRN8OtooU",
			Role: v1.Role_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_03802c004f5713db016a6c9a72decc81a09a8090abfd176c62",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJpzLJzi_lmKfuNMi51LbjdWpKSSgETp7r5XGu8ytH5JGPykbyT9G_SBFQA--cwvIyCEOICyoh6BFGBI_WwM-_NKcwtgl54UwSlBqsk7R87OLCKyuF9MqxGUVn2Q6A2NdynuJCKy-3-0oVMdelSuj6dJu_JhQokgPTd3t_y2uhEvJJguQmjFFjks0j-lFB-SiiXsW-FG2XHKT0DeTXDAILYAHnx9Fu-S2OFTwGzRxfj-TJZsmuqgn7b5tucAzaQYB7nAJz3uD35SRyRukzrZihahUR6PiLwILG3Ta3MXZSL8PSjwq0whLQgI0aQE-4-s1ZG9X2-uz54BQ-blp9g4lg5-WZlW2FxDC0dFCWAAqQthhw7oEHyfvJM6-cmbDs8zwgnUTKjtyGWABC1D97U0-Wh-dsHzpxAc3yszxHybDgW-jG2pjeUTEnZUAo26u8w0rnEO3qp8wLZY674if1P4NKk3JwJIhZhv78Fbq2bpMbW4afW8FJEnD2KojJesI-hAceIYj28CO5sGuCGN2a-gVjFIEiS3b5ThyMO9zxpI7-_phLVp7aE4cHqECGCULMDy3_dASm7HfkLCMQJMihQ3_yKEzZPfYIohae5JAgeYv2vCcJTkKaJASxM7V7_ejZGaXLNdPDTmIE68Y5o6GBrL6mO1Iesl-NJ8p7MN1vr1Rx2pONLizNCfUK5haYXT5XYyjDtxTqLA6Wk-aUbcJNB4_erODsNcaj5fJwFA1Vf2LlHVj5kx9xMFi0IogS3SZ6Rt72k8tGr1En7t_qve6Qp98zi1XQykVqe96sDYPyL42NgKNFotD-vSBobGeVdFjh2d_iKIoW-pnsRS8snUYitPS7Qzn38y7jJ8yInbKVB5BSJRIT1DCk1QoH1s-QUwEJWh0uQDEHbqwBdY_XeZIZWHHElxdShOYRG5w3y2dURNwm0-H9xowC5_zNXbq-6v3wVu3nbKq1QI1ZyDMQ5vnwRfyVpo7J5ANjSm3zo0oR-qeiHzJjg=",
					},
				},
				{Content: v1.NewTextContent("Neurouter successfully routed this Responses API request.")},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 38, OutputTokens: 28},
		},
	},
}
