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

func ContainerexecHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		url := fmt.Sprintf("%s/containers/%s/exec", cfg.BaseURL, id)
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
		var result models.IdResponse
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

func CreateContainerexecTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_containers_id_exec",
		mcp.WithDescription("Create an exec instance"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID or name of container")),
		mcp.WithBoolean("Privileged", mcp.Description("Input parameter: Runs the exec process with extended privileges.")),
		mcp.WithString("User", mcp.Description("Input parameter: The user, and optionally, group to run the exec process inside the container. Format is one of: `user`, `user:group`, `uid`, or `uid:gid`.")),
		mcp.WithBoolean("AttachStderr", mcp.Description("Input parameter: Attach to `stderr` of the exec command.")),
		mcp.WithBoolean("AttachStdout", mcp.Description("Input parameter: Attach to `stdout` of the exec command.")),
		mcp.WithArray("Cmd", mcp.Description("Input parameter: Command to run, as a string or array of strings.")),
		mcp.WithString("DetachKeys", mcp.Description("Input parameter: Override the key sequence for detaching a container. Format is a single character `[a-Z]` or `ctrl-<value>` where `<value>` is one of: `a-z`, `@`, `^`, `[`, `,` or `_`.")),
		mcp.WithBoolean("Tty", mcp.Description("Input parameter: Allocate a pseudo-TTY.")),
		mcp.WithBoolean("AttachStdin", mcp.Description("Input parameter: Attach to `stdin` of the exec command.")),
		mcp.WithArray("Env", mcp.Description("Input parameter: A list of environment variables in the form `[\"VAR=value\", ...]`.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ContainerexecHandler(cfg),
	}
}
