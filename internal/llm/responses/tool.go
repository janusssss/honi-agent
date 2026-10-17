package responses

import "encoding/json"

const (
	ToolTypeFunction = "function"
)

type Tool struct {
	Type        string          `json:"type"` // function
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema 原样透传
	Strict      bool            `json:"strict"`
}
