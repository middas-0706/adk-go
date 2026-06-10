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
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/agent"
)

func TestWorkflowConfig_StateSchemaConsistency(t *testing.T) {
	schemaRaw := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"Foo": {Type: "string"},
		},
	}
	schema, err := schemaRaw.Resolve(nil)
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	validNode, err := NewFunctionNodeFromState("valid_node", dummyFnValid, NodeConfig{})
	if err != nil {
		t.Fatalf("NewFunctionNodeFromState: %v", err)
	}

	invalidNode, err := NewFunctionNodeFromState("invalid_node", dummyFnInvalid, NodeConfig{})
	if err != nil {
		t.Fatalf("NewFunctionNodeFromState: %v", err)
	}

	agentNode := newDummyNode("agentNode")

	tests := []struct {
		name    string
		nodes   []Node
		schema  *jsonschema.Resolved
		wantErr string
	}{
		{
			name:   "valid schema matches exactly",
			nodes:  []Node{validNode},
			schema: schema,
		},
		{
			name:    "typo in tag causes error",
			nodes:   []Node{invalidNode},
			schema:  schema,
			wantErr: `node "invalid_node" references state field "foo" which is not declared in StateSchema`,
		},
		{
			name:   "nil schema skips validation",
			nodes:  []Node{invalidNode},
			schema: nil,
		},
		{
			name:   "non state params aware nodes are ignored",
			nodes:  []Node{agentNode},
			schema: schema,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			edges := fanOutFromStart(tc.nodes)
			_, err := New("wf", edges, WorkflowConfig{StateSchema: tc.schema})
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("expected error to contain %q, got: %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

type validParams struct {
	Foo       string `state:"Foo"`
	NodeInput string `state:"node_input"`
}

type invalidParams struct {
	Foo string `state:"foo"` // Typo: case mismatch
}

func dummyFnValid(ctx agent.InvocationContext, p validParams) (string, error) {
	return "ok", nil
}

func dummyFnInvalid(ctx agent.InvocationContext, p invalidParams) (string, error) {
	return "ok", nil
}
