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

func ImagelistHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["all"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("all=%v", val))
		}
		if val, ok := args["filters"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("filters=%v", val))
		}
		if val, ok := args["digests"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("digests=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		url := fmt.Sprintf("%s/images/json%s", cfg.BaseURL, queryString)
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
		var result []ImageSummary
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

func CreateImagelistTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("get_images_json",
		mcp.WithDescription("List Images"),
		mcp.WithBoolean("all", mcp.Description("Show all images. Only images from a final layer (no children) are shown by default.")),
		mcp.WithString("filters", mcp.Description("A JSON encoded value of the filters (a `map[string][]string`) to process on the images list. Available filters:\n\n- `before`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)\n- `dangling=true`\n- `label=key` or `label=\"key=value\"` of an image label\n- `reference`=(`<image-name>[:<tag>]`)\n- `since`=(`<image-name>[:<tag>]`,  `<image id>` or `<image@digest>`)\n")),
		mcp.WithBoolean("digests", mcp.Description("Show digest information as a `RepoDigests` field on each image.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ImagelistHandler(cfg),
	}
}
