// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"fmt"
	"slices"

	"github.com/google/jsonschema-go/jsonschema"
)

// WorkflowConfig configures the workflow at construction time.
type WorkflowConfig struct {
	StateSchema *jsonschema.Resolved
}

func validateStateSchemaConsistency(g *graph, schema *jsonschema.Resolved) error {
	if schema == nil {
		return nil
	}
	schemaFields := extractFieldNames(schema)

	for _, n := range g.allNodes() {
		spa, ok := n.(StateParamsAware)
		if !ok {
			continue
		}
		for _, fieldName := range spa.StateFieldNames() {
			if !slices.Contains(schemaFields, fieldName) {
				return fmt.Errorf("node %q references state field %q which is not declared in StateSchema (declared: %v)", n.Name(), fieldName, schemaFields)
			}
		}
	}
	return nil
}

func extractFieldNames(schema *jsonschema.Resolved) []string {
	var fields []string
	if schema != nil && schema.Schema() != nil && schema.Schema().Properties != nil {
		for k := range schema.Schema().Properties {
			fields = append(fields, k)
		}
	}
	slices.Sort(fields)
	return fields
}
