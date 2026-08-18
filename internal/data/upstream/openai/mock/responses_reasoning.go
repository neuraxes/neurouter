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

//go:embed responses_reasoning_request.json
var responsesReasoningRequest []byte

//go:embed responses_reasoning_response.json
var responsesReasoningResponse []byte

// ResponsesReasoning covers a reasoning summary, encrypted reasoning state and
// final text output.
var ResponsesReasoning = &Fixture{
	Name:     "responses_reasoning",
	Request:  responsesReasoningRequest,
	Response: responsesReasoningResponse,
	ChatRequest: &v1.ChatRequest{
		Id:    "responses_reasoning",
		Model: "openai/gpt-5-mini",
		Config: &v1.GenerationConfig{
			MaxTokens:       new(int64(2048)),
			ReasoningConfig: &v1.ReasoningConfig{Effort: v1.ReasoningEffort_REASONING_EFFORT_MEDIUM},
		},
		Messages: []*v1.Message{
			{
				Role: v1.Role_ROLE_SYSTEM,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("You are a conversion-test assistant. Solve the routing calculation carefully, then give a concise final answer.")},
				},
			},
			{
				Role: v1.Role_ROLE_USER,
				Contents: []*v1.Content{
					{Content: v1.NewTextContent("An LLM router estimates one-retry latency as L * (1 + e), where L is wait time in milliseconds and e is error rate. Compare OpenAI (L=1800, e=0.12), Anthropic (L=120, e=0.01), and Gemini (L=400, e=0.04). Calculate all three values and select the upstream with the lowest expected latency.")},
				},
			},
		},
	},
	ChatResponse: &v1.ChatResponse{
		Id:     "responses_reasoning",
		Model:  "openai/gpt-5-mini",
		Status: v1.ChatStatus_CHAT_STATUS_COMPLETED,
		Message: &v1.Message{
			Id:   "gen-1785502365-TGaPDCczhSq5BUQNfGH5",
			Role: v1.Role_ROLE_MODEL,
			Contents: []*v1.Content{
				{
					Id:    "rs_031a7050398ad760016a6c9a9e0234819d88eff125cf47f68f",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: v1.NewTextContent(`**Calculating one-retry latency**

I need to compute the one-retry latency using the formula L*(1+e) for three providers. For OpenAI, with L=1800 ms and e=0.12, it calculates to 2016 ms. Anthropic, with L=120 ms and e=0.01, results in 121.2 ms. Lastly, Gemini, with L=400 ms and e=0.04, gives 416 ms. The lowest latency is from Anthropic at 121.2 ms. I’ll make sure to present all calculations clearly and concisely.`),
				},
				{
					Id:    "rs_031a7050398ad760016a6c9a9e0234819d88eff125cf47f68f",
					Phase: v1.ContentPhase_CONTENT_PHASE_REASONING,
					Content: &v1.Content_Opaque{
						Opaque: "gAAAAABqbJqi7azdim4vT9BraH3QFRTJTzWB0BPT0lBcglWJLSbnK8wLQIITMXom8OWYGpVMPdeNE_oVWtzcj7ndxtjGYjszN6PZY_j9_SlDmaAlOSkEO4glgfPsoeKa6uXThbchIi0dSQcDfuTum1DVKI4-EuGLnT_sozP7SIo6ZBCXFWaFHaPYGhIOdZCRsnHqq-IO38MxtprDqkR4ssQqLhbJGVbUNvKBaBKTO8sR-xA4n2Lmyj5_6fXDvK_XDD95vhWf5FFcGSWKJJT_lLUlzSWdnJkgu-wCudAWuww6DGoyY9O-SQmb6OmmuAa_EKJ24BcTtsUi8BMqCN0xOrOtW6N9gYxssDqC_ceZfI5RJXGGOdMOK7T-y6ZZVgy2iisDsptsFaWU2cr56F9UyWVO2WD5r3ecgSIqS2yq1vlWRoi5U2PKktCsbSlas6PxwQSNb7DR1gl1fJMmpVnyb4hhGPlAKTT7zUdbivhkpJNvwyz1XHFlc4Bjpc90Ye1bbXT-qX7seMdmGnOHRNzB-QutEH34NuC_r4RlIVjwiwMyl1Ne0y1S1QsjntFA4q9lY5lEyD0Cc_y6Zgbfp7X72UqXUKcrz2dT9rA-k4SoT7toPOrnRfsveGfm5ZiGDCPG30W6utNGciB56zTHLE4LipL2CSH7YZ6bVJQgBU3DWxuDDisv8j58jQURPXK1PfGZwqp7clDIeYUl2_oqQTZX0kQ1fbGJT3aMYDhV54m80X2sGMBGiPMtXBr-3p5YEklAWNP0oFX8xMvbPNzTzFPoF40BFkj78hrj5ua4J20x1_Gr5ReUGQzl1iBbgow65807LEXgPWMypIrhCOp-vGy8XtBOiW-1LRsVGyA4iX41w7yZ2xTZXHRarcrF0qC1wy8TgYqjMikjzbkb3yyR9E0mCljjj867hH8czYMtfdm2kOstfj4alpHh4KpyNjF5GUyscPresvEjh7DZZ0QpitCpOeLTVjufsg9cgMRIzbY9pSK8guirHfK7Z_olXkKNRMvAhSlUyhbhRHfFw92_Ph6iU8qPfjw6Ig2vopPjsccxYKR2_pixKyKl98ipHsmio394bfaAuBvMK4BpplysL9BdTw7J4_SXImrrHS3ruEVW_7qIalNFIyO5Es7AZox_1Gevqp_wfhNwKyA203nRNOK-_A0eG_Wb0UZxg-l1jGQ_SR9DFttECQLl1Bl7Ds_no4OqWFAI30SQoBqVx1vuK78MbfUuxTqjZurRGizNf8Ir65_HYntPXZClu15lTa2ai-QFmGDUe7t_jKdC6QROlpUmw86GfbvAGMWz1e3nXnnVwV56Thohr_e2_THKWIU6etqg-j7Tt2rzhPnnNX-0ZDyAwJ9uoWhGGAsm97pOtUPwTqOcRkVlcn2vL4x0KUe0YzWnGcYwwHvmaeFFXXWeIo6urQPIc7ezGd12nRNMoe2CzhbfOG_WdeSc2kTK4udslHQbIlj4Cl62w350gs_lK4jPhmFdQYOZ0ypjm642NLUFa-ifGBDjUM9u2db5IhJWTvBmB2vaVrHsgNI2Jjq9sNJNCPXgFQ-ZVPgAqM6aynR6E5XhQXJNnK_Eo8oiwhYdu7lT6jEUSnnlCULtRH3Zasf1n7hITKtO8QGaVHP1uqnkzLvhajA_cn2QaX3O6EUKdGZqHXoKvm75iJ1ehaI6XHSxCPFtjEEKA5Wv-1dipGGFcm5dX0O7oImiPNjGpBE68kdk3SWOJCe7sis5z28DR_algMr4FssxvAEy9-KvoNaU4m1lPMUmj5EigLxFNpvw9uMon10daPcS5rUJjbjgdmi7qqaC6ThnhoRwXnTpdQJ-tQtaFvzCxOEvWPwh3MyYSAc6xaJzy7CW54faJpjHMxh93C8qqJBdQO4Y_EQcBf9Tk-wFcn9j9OtTdbgpA3VuzvvdY2iVJWLbTqmOjcqgLRe1_BnvlKGShNTmqnzDBoDc2UQw1-rH0fqrRhW8K9VsKbZ0Wa8ok_cGM5KvnfbQWnOrY1eLNC1vKHqrbfINLFIiut5Tun6Sm0U_G3M=",
					},
				},
				{
					Id: "msg_tmp_m6y3bvjpkrb",
					Content: v1.NewTextContent(`Compute L*(1+e):

- OpenAI: 1800 * (1 + 0.12) = 1800 * 1.12 = 2016 ms
- Anthropic: 120 * (1 + 0.01) = 120 * 1.01 = 121.2 ms
- Gemini: 400 * (1 + 0.04) = 400 * 1.04 = 416 ms

Lowest expected latency: Anthropic (121.2 ms).`),
				},
			},
		},
		Statistics: &v1.Statistics{
			Usage: &v1.Usage{InputTokens: 115, OutputTokens: 343, ReasoningTokens: 192},
		},
	},
}
