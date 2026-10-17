package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"honi/internal/llm/responses"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/zendev-sh/goai"
)

var GetMCPResource = sync.OnceValue(func() *MCPResource {
	m := &MCPResource{
		clients: []mcpClient{
			//&Anysearch{},
			&fileSystem{},
		},
		mu:      &sync.Mutex{},
		toolMap: make(map[string][]goai.Tool),
	}
	m.initTool()
	return m
})

type MCPResource struct {
	toolMap map[string][]goai.Tool
	active  []goai.Tool
	clients []mcpClient
	mu      *sync.Mutex
}

type mcpClient interface {
	tools() ([]goai.Tool, error)
	initialize() error
	getName() string
	getDescription() string
}

func (m *MCPResource) fetchTool() error {
	wg := sync.WaitGroup{}
	wg.Add(len(m.clients))
	for _, c := range m.clients {
		go func() {
			defer wg.Done()
			if err := c.initialize(); err != nil {
				slog.Error("connect to map server failed", "err", err)
				return
			}

			tools, err := c.tools()
			if err != nil {
				slog.Error("get tool from mcp resources failed", "err", err)
				return
			}

			m.mu.Lock()
			m.toolMap[c.getName()] = tools
			m.mu.Unlock()
		}()
	}

	wg.Wait()
	return nil
}

func (m *MCPResource) toolSetSchema() string {
	var schema strings.Builder
	for _, c := range m.clients {
		fmt.Fprintln(&schema, c.getName(), ":", c.getDescription())
	}
	return schema.String()
}

func (m *MCPResource) initTool() ([]goai.Tool, error) {
	m.fetchTool()
	_ = goai.NewTool(
		"",
		"加载资源，开放如下工具资源以供使用:\n"+m.toolSetSchema(),
		func(ctx context.Context, in struct {
			Name string `json:"name" jsonschema:"description=需要开启工具资源name"`
		}) (string, error) {
			tools := m.toolMap[in.Name]
			if len(tools) == 0 {
				return "", fmt.Errorf("找不到%v资源", in.Name)
			}
			m.active = append(m.active, tools...)
			return in.Name + "资源已启用", nil
		})

	for _, tools := range m.toolMap {
		m.active = append(m.active, tools...)
	}
	return m.active, nil
}

func (m *MCPResource) ActionTools() []goai.Tool {
	return m.active
}

const (
	toolTypeFunction = "function"
)

type Tool struct {
	tools []responses.Tool

	CallMap map[string]func(input json.RawMessage) (output json.RawMessage, err error) `json:"-"`
}
type schema struct {
	Type                 string                `json:"type"`
	Properties           map[string]properties `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	AdditionalProperties bool                  `json:"additionalProperties,omitempty"`
}
type properties struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (t *Tool) addToolByMcp(mcp mcpToolInterface) error {
	tools, err := mcp.tools()
	if err != nil {
		slog.Error("Tool addToolByMcp failed", "err", err)
	}

	t.tools = append(t.tools, tools...)

	for _, tool := range tools {
		t.CallMap[tool.Name] = func(input json.RawMessage) (output json.RawMessage, err error) {
			return mcp.callTool(tool.Name, input)
		}
	}
	return nil
}

func (t *Tool) addTool[T any](name, description string, input T,
	call func(input json.RawMessage) (output json.RawMessage, err error)) (
	err error) {
	vo := reflect.ValueOf(input)
	if vo.Kind() != reflect.Struct {
		return errors.New("input isn't a struct")
	}

	_schema := schema{
		Type: "object",
	}
	for i := 0; i < vo.NumField(); i++ {
		vf := vo.Field(i)
		vt := vf.Type()
		tag := vt.Field(i).Tag
		name, ok := tag.Lookup("json")
		if !ok {
			name = vt.Name()
		}
		description, ok := tag.Lookup("description")
		if !ok {
			return errors.New("newnewTool inpot field no tag dedescription")
		}
		_schema.Properties[name] = properties{
			Type:        vt.String(),
			Description: description,
		}
	}

	buff, err := json.Marshal(_schema)
	if err != nil {
		slog.Error("newTool Marshal failed", "err", err)
		return err
	}

	t.tools = append(t.tools,
		responses.Tool{
			Type:        toolTypeFunction,
			Name:        name,
			Description: description,
			Parameters:  buff,
		})
	t.CallMap[name] = call
	return
}

var GetTool = sync.OnceValue(func() *Tool {
	t := &Tool{
		CallMap: map[string]func(input json.RawMessage) (output json.RawMessage, err error){},
	}
	if err := t.addToolByMcp(getAnysearch()); err != nil {
		slog.Warn("GetTool addToolByMcp failed", "err", err)
	}
	return t
})

func (t *Tool) ActiveTools() []responses.Tool {
	return t.tools
}

func (t *Tool) CallTool(name string, input json.RawMessage) (answer json.RawMessage, err error) {
	return t.CallMap[name](input)
}
