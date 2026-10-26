package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestBodyValidation(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   string
	}{
		{"project without name", http.MethodPost, "/api/v1/projects", `{}`,
			"invalid body: name is required"},
		{"section unknown scope", http.MethodPost, "/api/v1/sections", `{"scope":"team","name":"style"}`,
			"invalid body: scope must be one of: global, project"},
		{"section all missing", http.MethodPost, "/api/v1/sections", `{}`,
			"invalid body: scope is required; name is required"},
		{"section project scope without project", http.MethodPost, "/api/v1/sections", `{"scope":"project","name":"style"}`,
			"invalid body: project is required for scope project"},
		{"section global scope with project", http.MethodPost, "/api/v1/sections", `{"scope":"global","project":"app","name":"style"}`,
			"invalid body: project is not allowed for scope global"},
		{"context without section and name", http.MethodPost, "/api/v1/contexts", `{"scope":"global"}`,
			"invalid body: section is required; name is required"},
		{"context project scope without project", http.MethodPost, "/api/v1/contexts", `{"scope":"project","section":"style","name":"go"}`,
			"invalid body: project is required for scope project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newTestServer(t)
			rec := do(t, h, tt.method, tt.path, tt.body)
			expectStatus(t, rec, http.StatusBadRequest)
			if got := decodeBody[errorResponse](t, rec).Error; got != tt.want {
				t.Errorf("error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateContextEmptyBodyIsValid(t *testing.T) {
	h, _ := newTestServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", `{"scope":"global","section":"style","name":"go"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPut, "/api/v1/contexts/global/style/go", `{}`), http.StatusOK)
}

func TestSwaggerOperationIDs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "swagger", "swagger.json"))
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}
	var spec struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("decode swagger.json: %v", err)
	}
	seen := map[string]string{}
	for path, ops := range spec.Paths {
		for method, op := range ops {
			where := method + " " + path
			if op.OperationID == "" {
				t.Errorf("%s: missing operationId", where)
				continue
			}
			if prev, ok := seen[op.OperationID]; ok {
				t.Errorf("%s: operationId %q also used by %s", where, op.OperationID, prev)
			}
			seen[op.OperationID] = where
		}
	}
}
