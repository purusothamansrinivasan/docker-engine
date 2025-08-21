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

func SystemeventsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["since"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("since=%v", val))
		}
		if val, ok := args["until"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("until=%v", val))
		}
		if val, ok := args["filters"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filters=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/events%s", cfg.BaseURL, queryString)
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

func CreateSystemeventsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_events",
		mcp.WithDescription("Monitor events"),
		mcp.WithString("since", mcp.Description("Show events created since this timestamp then stream new events.")),
		mcp.WithString("until", mcp.Description("Show events created until this timestamp then stop streaming.")),
		mcp.WithString("filters", mcp.Description("A JSON encoded value of filters (a `map[string][]string`) to process on the event list. Available filters:\n\n- `config=<string>` config name or ID\n- `container=<string>` container name or ID\n- `daemon=<string>` daemon name or ID\n- `event=<string>` event type\n- `image=<string>` image name or ID\n- `label=<string>` image or container label\n- `network=<string>` network name or ID\n- `node=<string>` node ID\n- `plugin`=<string> plugin name or ID\n- `scope`＝<string> local or swarm\n- `secret=<string>` secret name or ID\n- `service=<string>` service name or ID\n- `type=<string>` object to filter by, one of `container`, `image`, `volume`, `network`, `daemon`, `plugin`, `node`, `service`, `secret` or `config`\n- `volume=<string>` volume name\n")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    SystemeventsHandler(cfg),
	}
}
