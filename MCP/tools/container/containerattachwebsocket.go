package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker-engine-api/mcp-server/config"
	"github.com/docker-engine-api/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func ContainerattachwebsocketHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		idVal, ok := args["id"]
		if !ok {
			return mcp.NewToolResultError("Missing required path parameter: id"), nil
		}
		id, ok := idVal.(string)
		if !ok {
			return mcp.NewToolResultError("Invalid path parameter: id"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["detachKeys"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("detachKeys=%v", val))
		}
		if val, ok := args["logs"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("logs=%v", val))
		}
		if val, ok := args["stream"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stream=%v", val))
		}
		if val, ok := args["stdin"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stdin=%v", val))
		}
		if val, ok := args["stdout"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stdout=%v", val))
		}
		if val, ok := args["stderr"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stderr=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/containers/%s/attach/ws%s", cfg.BaseURL, id, queryString)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create request", err), nil
		}
		// No authentication required for this endpoint
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Request failed", err), nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to read response body", err), nil
		}

		if resp.StatusCode >= 400 {
			return mcp.NewToolResultError(fmt.Sprintf("API error: %s", body)), nil
		}
		// Use properly typed response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			// Fallback to raw text if unmarshaling fails
			return mcp.NewToolResultText(string(body)), nil
		}

		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to format JSON", err), nil
		}

		return mcp.NewToolResultText(string(prettyJSON)), nil
	}
}

func CreateContainerattachwebsocketTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_containers_id_attach_ws",
		mcp.WithDescription("Attach to a container via a websocket"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID or name of the container")),
		mcp.WithString("detachKeys", mcp.Description("Override the key sequence for detaching a container.Format is a single character `[a-Z]` or `ctrl-<value>` where `<value>` is one of: `a-z`, `@`, `^`, `[`, `,`, or `_`.")),
		mcp.WithBoolean("logs", mcp.Description("Return logs")),
		mcp.WithBoolean("stream", mcp.Description("Return stream")),
		mcp.WithBoolean("stdin", mcp.Description("Attach to `stdin`")),
		mcp.WithBoolean("stdout", mcp.Description("Attach to `stdout`")),
		mcp.WithBoolean("stderr", mcp.Description("Attach to `stderr`")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ContainerattachwebsocketHandler(cfg),
	}
}
