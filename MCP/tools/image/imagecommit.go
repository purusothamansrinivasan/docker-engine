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

func ImagecommitHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments object"), nil
		}
		queryParams := make([]string, 0)
		if val, ok := args["container"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("container=%v", val))
		}
		if val, ok := args["repo"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("repo=%v", val))
		}
		if val, ok := args["tag"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("tag=%v", val))
		}
		if val, ok := args["comment"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("comment=%v", val))
		}
		if val, ok := args["author"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("author=%v", val))
		}
		if val, ok := args["pause"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("pause=%v", val))
		}
		if val, ok := args["changes"]; ok {
			queryParams = append(queryParams, fmt.Sprintf("changes=%v", val))
		}
		queryString := ""
		if len(queryParams) > 0 {
			queryString = "?" + strings.Join(queryParams, "&")
		}
		// Create properly typed request body using the generated schema
		var requestBody models.ContainerConfig
		
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
		url := fmt.Sprintf("%s/commit%s", cfg.BaseURL, queryString)
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

func CreateImagecommitTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_commit",
		mcp.WithDescription("Create a new image from a container"),
		mcp.WithString("container", mcp.Description("The ID or name of the container to commit")),
		mcp.WithString("repo", mcp.Description("Repository name for the created image")),
		mcp.WithString("tag", mcp.Description("Tag name for the create image")),
		mcp.WithString("comment", mcp.Description("Commit message")),
		mcp.WithString("author", mcp.Description("Author of the image (e.g., `John Hannibal Smith <hannibal@a-team.com>`)")),
		mcp.WithBoolean("pause", mcp.Description("Whether to pause the container before committing")),
		mcp.WithString("changes", mcp.Description("`Dockerfile` instructions to apply while committing")),
		mcp.WithString("Hostname", mcp.Description("Input parameter: The hostname to use for the container, as a valid RFC 1123 hostname.")),
		mcp.WithObject("Labels", mcp.Description("Input parameter: User-defined key/value metadata.")),
		mcp.WithBoolean("ArgsEscaped", mcp.Description("Input parameter: Command is already escaped (Windows only)")),
		mcp.WithNumber("StopTimeout", mcp.Description("Input parameter: Timeout to stop a container in seconds.")),
		mcp.WithObject("Volumes", mcp.Description("Input parameter: An object mapping mount point paths inside the container to empty objects.")),
		mcp.WithBoolean("AttachStdout", mcp.Description("Input parameter: Whether to attach to `stdout`.")),
		mcp.WithArray("Env", mcp.Description("Input parameter: A list of environment variables to set inside the container in the form `[\"VAR=value\", ...]`. A variable without `=` is removed from the environment, rather than to have an empty value.\n")),
		mcp.WithString("StopSignal", mcp.Description("Input parameter: Signal to stop a container as a string or unsigned integer.")),
		mcp.WithString("Entrypoint", mcp.Description("Input parameter: The entry point for the container as a string or an array of strings.\n\nIf the array consists of exactly one empty string (`[\"\"]`) then the entry point is reset to system default (i.e., the entry point used by docker when there is no `ENTRYPOINT` instruction in the `Dockerfile`).\n")),
		mcp.WithString("Image", mcp.Description("Input parameter: The name of the image to use when creating the container")),
		mcp.WithBoolean("AttachStderr", mcp.Description("Input parameter: Whether to attach to `stderr`.")),
		mcp.WithArray("OnBuild", mcp.Description("Input parameter: `ONBUILD` metadata that were defined in the image's `Dockerfile`.")),
		mcp.WithString("WorkingDir", mcp.Description("Input parameter: The working directory for commands to run in.")),
		mcp.WithString("Cmd", mcp.Description("Input parameter: Command to run specified as a string or an array of strings.")),
		mcp.WithObject("Healthcheck", mcp.Description("Input parameter: A test to perform to check that the container is healthy.")),
		mcp.WithBoolean("StdinOnce", mcp.Description("Input parameter: Close `stdin` after one attached client disconnects")),
		mcp.WithObject("ExposedPorts", mcp.Description("Input parameter: An object mapping ports to an empty object in the form:\n\n`{\"<port>/<tcp|udp>\": {}}`\n")),
		mcp.WithString("Domainname", mcp.Description("Input parameter: The domain name to use for the container.")),
		mcp.WithBoolean("NetworkDisabled", mcp.Description("Input parameter: Disable networking for the container.")),
		mcp.WithBoolean("AttachStdin", mcp.Description("Input parameter: Whether to attach to `stdin`.")),
		mcp.WithArray("Shell", mcp.Description("Input parameter: Shell for when `RUN`, `CMD`, and `ENTRYPOINT` uses a shell.")),
		mcp.WithBoolean("Tty", mcp.Description("Input parameter: Attach standard streams to a TTY, including `stdin` if it is not closed.")),
		mcp.WithString("MacAddress", mcp.Description("Input parameter: MAC address of the container.")),
		mcp.WithBoolean("OpenStdin", mcp.Description("Input parameter: Open `stdin`")),
		mcp.WithString("User", mcp.Description("Input parameter: The user that commands are run as inside the container.")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ImagecommitHandler(cfg),
	}
}
