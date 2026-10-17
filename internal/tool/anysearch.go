package tool

import (
	"context"
	"encoding/json"
	"errors"
	"honi/config"
	"honi/internal/llm/responses"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	gmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	anysearchName        = "anysearch"
	anysearchDescription = `
Unified real-time search engine skill for AI agents.Supports general web search,
vertical domain search, parallel batch search,and full-page content extraction.`
)

type Anysearch struct {
	session *gmcp.ClientSession
}

var getAnysearch = sync.OnceValue(func() *Anysearch {
	anysearch := &Anysearch{}
	session, err := anysearch.connect()
	if err != nil {
		slog.Warn("getAnysearch Connect failed", "err", err)
	}
	anysearch.session = session
	return anysearch
})

func (a *Anysearch) name() string {
	return anysearchName
}

func (a *Anysearch) description() string {
	return strings.TrimSpace(anysearchDescription)
}

func (a *Anysearch) tools() (tools []responses.Tool, err error) {
	toolResult, err := a.session.ListTools(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	for _, tool := range toolResult.Tools {
		input, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, err
		}
		tool := responses.Tool{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  input,
		}
		tools = append(tools, tool)
	}

	return
}

func (a *Anysearch) callTool(name string, input json.RawMessage) (output json.RawMessage, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	args := map[string]any{}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		slog.Error("Anysearch Unmarshal failed", "err", err)
		return nil, err
	}
	callResult, err := a.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		slog.Error("Anysearch CallTool failed", "err", err)
		return output, err
	}

	var result [][]byte
	for _, content := range callResult.Content {
		content, err := content.MarshalJSON()
		if err != nil {
			return output, err
		}
		result = append(result, content)
	}

	output, err = json.Marshal(result)
	if err != nil {
		slog.Error("Anysearch Marshal failed", "err", err)
	}
	return
}

func (a *Anysearch) connect() (*gmcp.ClientSession, error) {
	ctx := context.Background()
	client := gmcp.NewClient(&gmcp.Implementation{Name: anysearchName, Version: "v1.0.0"}, nil)

	mcps := config.GetConfig().MCPs
	index := slices.IndexFunc(mcps, func(mcp config.MCP) bool {
		return mcp.Name == anysearchName
	})
	if index < 0 {
		return nil, errors.New("Anysearch map not config")
	}
	mcp := mcps[index]

	session, err := client.Connect(ctx, &gmcp.StreamableClientTransport{
		Endpoint: mcp.Endpoint,
		HTTPClient: &http.Client{
			Transport: withAuth("Bearer " + mcp.ApiKey),
		},
	}, nil)
	if err != nil {
		return nil, err
	}
	return session, err
}

type authRT struct{ token string }

func withAuth(token string) http.RoundTripper {
	return &authRT{token: token}
}

func (a *authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", a.token)
	return http.DefaultTransport.RoundTrip(r)
}
