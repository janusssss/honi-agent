package responses

type FunctionCallOutput struct {
	CallId string `json:"call_id"`
	Output string `json:"output"`
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Tyep   string `json:"type"`
}
