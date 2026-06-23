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

package util

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func StructFromAny(value any) (*structpb.Struct, error) {
	if value == nil {
		return nil, nil
	}
	if schema, ok := value.(*structpb.Struct); ok {
		return schema, nil
	}
	if schema, ok := value.(map[string]any); ok {
		return StructFromMap(schema)
	}

	return structThroughJSON(value)
}

func StructFromMap(fields map[string]any) (*structpb.Struct, error) {
	if fields == nil {
		return nil, nil
	}
	if schema, err := structpb.NewStruct(fields); err == nil {
		return schema, nil
	}
	return structThroughJSON(fields)
}

func MustStructFromMap(fields map[string]any) *structpb.Struct {
	schema, err := StructFromMap(fields)
	if err != nil {
		panic(err)
	}
	return schema
}

func structThroughJSON(value any) (*structpb.Struct, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || data[0] != '{' {
		return nil, fmt.Errorf("value must be a JSON object")
	}
	schema := &structpb.Struct{}
	if err := protojson.Unmarshal(data, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func StringSliceFromAny(value any) ([]string, bool) {
	switch v := value.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, s)
		}
		return result, true
	default:
		return nil, false
	}
}
