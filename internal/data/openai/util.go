package openai

import (
	"github.com/openai/openai-go"

	v1 "git.xdea.xyz/Turing/router/api/laas/v1"
)

func toolFunctionParametersToOpenAI(parameters *v1.Tool_Function_Parameters) openai.FunctionParameters {
	return map[string]any{
		"type":       parameters.Type,
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}
