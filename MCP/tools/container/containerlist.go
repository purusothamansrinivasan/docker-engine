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

func ContainerlistHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["all"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("all=%v", val))
		}
		if val, ok := args["limit"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("limit=%v", val))
		}
		if val, ok := args["size"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("size=%v", val))
		}
		if val, ok := args["filters"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filters=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/containers/json%s", cfg.BaseURL, queryString)
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
		var result []map[string]interface{}
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

func CreateContainerlistTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_containers_json",
		mcp.WithDescription("List containers"),
		mcp.WithBoolean("all", mcp.Description("Return all containers. By default, only running containers are shown")),
		mcp.WithNumber("limit", mcp.Description("Return this number of most recently created containers, including non-running ones.")),
		mcp.WithBoolean("size", mcp.Description("Return the size of container as fields `SizeRw` and `SizeRootFs`.")),
		mcp.WithString("filters", mcp.Description("Filters to process on the container list, encoded as JSON (a `map[string][]string`). For example, `{\"status\": [\"paused\"]}` will only return paused containers. Available filters:\n\n- `ancestor`=(`<image-name>[:<tag>]`, `<image id>`, or `<image@digest>`)\n- `before`=(`<container id>` or `<container name>`)\n- `expose`=(`<port>[/<proto>]`|`<startport-endport>/[<proto>]`)\n- `exited=<int>` containers with exit code of `<int>`\n- `health`=(`starting`|`healthy`|`unhealthy`|`none`)\n- `id=<ID>` a container's ID\n- `isolation=`(`default`|`process`|`hyperv`) (Windows daemon only)\n- `is-task=`(`true`|`false`)\n- `label=key` or `label=\"key=value\"` of a container label\n- `name=<name>` a container's name\n- `network`=(`<network id>` or `<network name>`)\n- `publish`=(`<port>[/<proto>]`|`<startport-endport>/[<proto>]`)\n- `since`=(`<container id>` or `<container name>`)\n- `status=`(`created`|`restarting`|`running`|`removing`|`paused`|`exited`|`dead`)\n- `volume`=(`<volume name>` or `<mount point destination>`)\n")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ContainerlistHandler(cfg),
	}
}
