package tool

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAnysearchTool(t *testing.T) {
	var c mcpToolInterface = &Anysearch{}
	tools, err := c.tools()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", tools)
}

func TestGetName(t *testing.T) {
	var c mcpToolInterface = &Anysearch{}
	require.Equal(t, "anysearch", c.name())
}

func TestAnysearchToolByMcp(t *testing.T) {
	as := Anysearch{}
	session, err := as.connect()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		fmt.Println(tool.Name)
		fmt.Println(tool.InputSchema)
	}
}

func TestAnysearchCall(t *testing.T) {
	as := Anysearch{}
	session, err := as.connect()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	args := map[string]any{
		"query":       "中国哪个女人最美",
		"max_results": 5, // float64/int 都行，schema 是 number
	}
	reslut, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "search", Arguments: args})
	if err != nil {
		t.Fatal(err)
	}

	for _, content := range reslut.Content {
		buff, err := content.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(buff))
	}
}
