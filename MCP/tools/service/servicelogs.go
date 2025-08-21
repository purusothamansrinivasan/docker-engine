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

func ServicelogsHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		if val, ok := args["details"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("details=%v", val))
		}
		if val, ok := args["follow"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("follow=%v", val))
		}
		if val, ok := args["stdout"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stdout=%v", val))
		}
		if val, ok := args["stderr"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("stderr=%v", val))
		}
		if val, ok := args["since"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("since=%v", val))
		}
		if val, ok := args["timestamps"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("timestamps=%v", val))
		}
		if val, ok := args["tail"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("tail=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/services/%s/logs%s", cfg.BaseURL, id, queryString)
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
		var result string
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

func CreateServicelogsTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_services_id_logs",
		mcp.WithDescription("Get service logs"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID or name of the service")),
		mcp.WithBoolean("details", mcp.Description("Show service context and extra details provided to logs.")),
		mcp.WithBoolean("follow", mcp.Description("Return the logs as a stream.\n\nThis will return a `101` HTTP response with a `Connection: upgrade` header, then hijack the HTTP connection to send raw output. For more information about hijacking and the stream format, [see the documentation for the attach endpoint](#operation/ContainerAttach).\n")),
		mcp.WithBoolean("stdout", mcp.Description("Return logs from `stdout`")),
		mcp.WithBoolean("stderr", mcp.Description("Return logs from `stderr`")),
		mcp.WithNumber("since", mcp.Description("Only return logs since this time, as a UNIX timestamp")),
		mcp.WithBoolean("timestamps", mcp.Description("Add timestamps to every log line")),
		mcp.WithString("tail", mcp.Description("Only return this number of log lines from the end of the logs. Specify as an integer or `all` to output all log lines.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ServicelogsHandler(cfg),
	}
}
