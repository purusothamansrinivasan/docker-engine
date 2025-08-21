package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"bytes"

	"github.com/docker-engine-api/mcp-server/config"
	"github.com/docker-engine-api/mcp-server/models"
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceupdateHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		if val, ok := args["version"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("version=%v", val))
		}
		if val, ok := args["registryAuthFrom"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("registryAuthFrom=%v", val))
		}
		if val, ok := args["rollback"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("rollback=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
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
		url := fmt.Sprintf("%s/services/%s/update%s", cfg.BaseURL, id, queryString)
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
		var result models.ServiceUpdateResponse
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

func CreateServiceupdateTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_services_id_update",
		mcp.WithDescription("Update a service"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID or name of service.")),
		mcp.WithNumber("version", mcp.Required(), mcp.Description("The version number of the service object being updated. This is required to avoid conflicting writes.")),
		mcp.WithString("registryAuthFrom", mcp.Description("If the X-Registry-Auth header is not specified, this parameter indicates where to find registry authorization credentials. The valid values are `spec` and `previous-spec`.")),
		mcp.WithString("rollback", mcp.Description("Set to this parameter to `previous` to cause a server-side rollback to the previous service spec. The supplied spec will be ignored in this case.")),
		mcp.WithString("X-Registry-Auth", mcp.Description("A base64-encoded auth configuration for pulling from private registries. [See the authentication section for details.](#section/Authentication)")),
		mcp.WithObject("EndpointSpec", mcp.Description("Input parameter: Properties that can be configured to access and load balance a service.")),
		mcp.WithObject("Labels", mcp.Description("Input parameter: User-defined key/value metadata.")),
		mcp.WithObject("Mode", mcp.Description("Input parameter: Scheduling mode for the service.")),
		mcp.WithString("Name", mcp.Description("Input parameter: Name of the service.")),
		mcp.WithArray("Networks", mcp.Description("Input parameter: Array of network names or IDs to attach the service to.")),
		mcp.WithObject("RollbackConfig", mcp.Description("Input parameter: Specification for the rollback strategy of the service.")),
		mcp.WithObject("TaskTemplate", mcp.Description("Input parameter: User modifiable task configuration.")),
		mcp.WithObject("UpdateConfig", mcp.Description("Input parameter: Specification for the update strategy of the service.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ServiceupdateHandler(cfg),
	}
}
