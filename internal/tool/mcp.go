package tool

import (
	"encoding/json"
	"honi/internal/llm/responses"
	"sync"
)

// ===== MCP Tool 定义（tools/list 返回项） =====

type mcpTool struct {
	Name         string             `json:"name"`                   // 工具唯一标识
	Title        string             `json:"title,omitempty"`        // 供人阅读的展示名
	Description  string             `json:"description,omitempty"`  // 供模型理解的功能描述
	InputSchema  mcpSchema          `json:"inputSchema"`            // JSON Schema，即入参定义
	OutputSchema *mcpSchema         `json:"outputSchema,omitempty"` // 可选，结构化输出的校验 schema
	Annotations  mcpToolAnnotations `json:"annotations,omitempty"`  // 行为提示，不可全信
}

// mcpSchema 就是标准 JSON Schema，这里原样透传。
type mcpSchema = json.RawMessage

type mcpToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`    // 只读，不产生副作用
	DestructiveHint bool   `json:"destructiveHint,omitempty"` // 可能产生破坏性副作用
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`  // 重复调用结果一致
	OpenWorldHint   bool   `json:"openWorldHint,omitempty"`   // 与外部开放世界交互
}

// ===== MCP Tool 调用（tools/call） =====

type mcpToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"` // 需符合 inputSchema
}

type mcpToolResult struct {
	Content           []mcpContentBlock `json:"content"`                     // 非结构化内容块
	StructuredContent json.RawMessage   `json:"structuredContent,omitempty"` // 结构化输出，需符合 outputSchema
	IsError           bool              `json:"isError,omitempty"`           // true 表示执行失败，非协议错误
}

// mcpContentBlock 是 text | image | audio | resource_link | resource 的联合类型。
type mcpContentBlock struct {
	Type        string          `json:"type"`
	Text        string          `json:"text,omitempty"`        // text
	Data        string          `json:"data,omitempty"`        // image / audio：base64
	MimeType    string          `json:"mimeType,omitempty"`    // image / audio / resource_link
	URI         string          `json:"uri,omitempty"`         // resource_link / resource
	Name        string          `json:"name,omitempty"`        // resource_link
	Description string          `json:"description,omitempty"` // resource_link
	Resource    *mcpResource    `json:"resource,omitempty"`    // resource：内嵌资源
	Annotations *mcpAnnotations `json:"annotations,omitempty"` // 受众、优先级等元数据
}

type mcpResource struct {
	URI         string          `json:"uri"`
	MimeType    string          `json:"mimeType,omitempty"`
	Text        string          `json:"text,omitempty"`
	Blob        string          `json:"blob,omitempty"`
	Annotations *mcpAnnotations `json:"annotations,omitempty"`
}

type mcpAnnotations struct {
	Audience     []string `json:"audience,omitempty"` // user | assistant
	Priority     float64  `json:"priority,omitempty"`
	LastModified string   `json:"lastModified,omitempty"`
}

// ===== tools/list 响应 =====

type mcpListToolsResult struct {
	Tools      []mcpTool `json:"tools"`
	NextCursor string    `json:"nextCursor,omitempty"` // 非空表示有下一页
}

var GetMcp = sync.OnceValue(func() *Mcp {
	return &Mcp{}
})

type Mcp struct {
	tools []Tool
}

func (m *Mcp) Tools() []Tool {
	return m.tools
}

func FetchTool() (tools []Tool, err error) {
	/*
		type input struct {
			Name string `json:"name" description:"需要开启资源name"`
		}
		anysearch := &Anysearch{}
		session, err := anysearch.Connect()
		if err != nil {
			slog.Error("FetchTool anysearch connect failed", "err", err)
			return nil, err
		}
		toolResult, err := session.ListTools(context.Background(), nil)
		if err != nil {
			slog.Error("FetchTool ListTools failed", "err", err)
			return nil, err
		}
		var anysearchTools []Tool
		for _, tool := range toolResult.Tools {
			input, err := json.Marshal(tool.InputSchema)
			if err != nil {
				slog.Error("FetchTool Marshal failed", "err", err)
				return nil, err
			}
			rtool := Tool{
				Type:        responses.ToolTypeFunction,
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  input,
				Call: func(input json.RawMessage) (output []json.RawMessage, err error) {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()

					callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
						Name:      tool.Name,
						Arguments: input,
					})
					if err != nil {
						slog.Error("FetchTool CallTool failed", "err", err)
						return output, err
					}
					for _, content := range callResult.Content {
						content, err := content.MarshalJSON()
						if err != nil {
							return output, err
						}
						item := responses.FunctionCallOutput{
							CallId: "",
							Status: callResult.RequestState,
							Output: string(content),
							Tyep:   "function_call_output",
						}
						buff, err := json.Marshal(item)
						if err != nil {
							return output, nil
						}
						output = append(output, buff)
					}

					return
				},
			}
			anysearchTools = append(anysearchTools, rtool)
		}

		var tool Tool
		_input := input{}
		tool, err = newTool("加载工具资源", "anysearch:加载网页网页搜索的工具", _input,
			func(input json.RawMessage) (output json.RawMessage, err error) {
				if err = json.Unmarshal([]byte(input), &_input); err != nil {
					slog.Error("FetchTool Unmarshal failed", "err", err)
					return nil, err
				}
				mcp := GetMcp()
				mcp.tools = append(mcp.tools, tool)
				return []byte("加载工具资源完成"), nil
			})
		if err != nil {
			slog.Error("FetchTool newTool failed", "err", err)
			return nil, err
		}
		tools = append(tools, tool)
	*/
	return
}

type mcpToolInterface interface {
	name() string
	description() string
	tools() (tools []responses.Tool, err error)
	callTool(name string, input json.RawMessage) (output json.RawMessage, err error)
}
