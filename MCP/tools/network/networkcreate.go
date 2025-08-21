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

func NetworkcreateHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		// Create properly typed request body using the generated schema
		var requestBody map[string]interface{}
		
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
		url := fmt.Sprintf("%s/networks/create", cfg.BaseURL)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
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

func CreateNetworkcreateTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_networks_create",
		mcp.WithDescription("Create a network"),
		mcp.WithBoolean("Internal", mcp.Description("Input parameter: Restrict external access to the network.")),
		mcp.WithObject("Options", mcp.Description("Input parameter: Network specific options to be used by the drivers.")),
		mcp.WithBoolean("Attachable", mcp.Description("Input parameter: Globally scoped network is manually attachable by regular containers from workers in swarm mode.")),
		mcp.WithString("Driver", mcp.Description("Input parameter: Name of the network driver plugin to use.")),
		mcp.WithObject("IPAM", mcp.Description("")),
		mcp.WithBoolean("Ingress", mcp.Description("Input parameter: Ingress network is the network which provides the routing-mesh in swarm mode.")),
		mcp.WithString("Name", mcp.Required(), mcp.Description("Input parameter: The network's name.")),
		mcp.WithBoolean("CheckDuplicate", mcp.Description("Input parameter: Check for networks with duplicate names. Since Network is primarily keyed based on a random ID and not on the name, and network name is strictly a user-friendly alias to the network which is uniquely identified using ID, there is no guaranteed way to check for duplicates. CheckDuplicate is there to provide a best effort checking of any networks which has the same name but it is not guaranteed to catch all name collisions.")),
		mcp.WithObject("Labels", mcp.Description("Input parameter: User-defined key/value metadata.")),
		mcp.WithBoolean("EnableIPv6", mcp.Description("Input parameter: Enable IPv6 on the network.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    NetworkcreateHandler(cfg),
	}
}
