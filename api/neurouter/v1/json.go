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

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (x Schema_Type) MarshalJSON() ([]byte, error) {
	s, ok := Schema_Type_name[int32(x)]
	if !ok {
		return nil, fmt.Errorf("invalid schema type: %d", x)
	}
	return json.Marshal(strings.ToLower(strings.TrimPrefix(s, "TYPE_")))
}

func (x *Schema_Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("schema type should be a string, got %s", data)
	}
	v, ok := Schema_Type_value["TYPE_"+strings.ToUpper(s)]
	if !ok {
		return fmt.Errorf("invalid schema type %q", s)
	}
	*x = Schema_Type(v)
	return nil
}
