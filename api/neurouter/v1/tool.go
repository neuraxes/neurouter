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

func (x *Tool_Custom) GrammarDefinition() (syntax, definition string) {
	switch grammar := x.GetGrammar().(type) {
	case *Tool_Custom_LarkGrammar:
		return "lark", grammar.LarkGrammar
	case *Tool_Custom_RegexGrammar:
		return "regex", grammar.RegexGrammar
	default:
		return "", ""
	}
}
