package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"

	"github.com/docker-engine-api/mcp-server/config"
	"github.com/docker-engine-api/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func ServicecreateHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		// Create properly typed request body using the generated schema
		var requestBody interface{}
		
		// Optimized: Single marshal/unmarshal with JSON tags handling field mapping
		if argsJSON, err := json.Marshal(args); err == nil {
			if err := json.Unmarshal(argsJSON, &requestBody); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to convert arguments to request type: %v", err)), nil
			}
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal arguments: %v", err)), nil
		}
		
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to encode request body", err), nil
		}
		url := fmt.Sprintf("%s/services/create", cfg.BaseURL)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to create request", err), nil
		}
		// No authentication required for this endpoint
		req.Header.Set("Accept", "application/json")
		if val, ok := args["X-Registry-Auth"]; ok {
			req.Header.Set("X-Registry-Auth", fmt.Sprintf("%v", val))
		}

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

func CreateServicecreateTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_services_create",
		mcp.WithDescription("Create a service"),
		mcp.WithString("X-Registry-Auth", mcp.Description("A base64-encoded auth configuration for pulling from private registries. [See the authentication section for details.](#section/Authentication)")),
		mcp.WithObject("TaskTemplate", mcp.Description("Input parameter: User modifiable task configuration.")),
		mcp.WithObject("UpdateConfig", mcp.Description("Input parameter: Specification for the update strategy of the service.")),
		mcp.WithObject("EndpointSpec", mcp.Description("Input parameter: Properties that can be configured to access and load balance a service.")),
		mcp.WithObject("Labels", mcp.Description("Input parameter: User-defined key/value metadata.")),
		mcp.WithObject("Mode", mcp.Description("Input parameter: Scheduling mode for the service.")),
		mcp.WithString("Name", mcp.Description("Input parameter: Name of the service.")),
		mcp.WithArray("Networks", mcp.Description("Input parameter: Array of network names or IDs to attach the service to.")),
		mcp.WithObject("RollbackConfig", mcp.Description("Input parameter: Specification for the rollback strategy of the service.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ServicecreateHandler(cfg),
	}
}
