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

func ContainerupdateHandler(cfg *config.APIConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		url := fmt.Sprintf("%s/containers/%s/update", cfg.BaseURL, id)
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

func CreateContainerupdateTool(cfg *config.APIConfig) models.Tool {
	tool := mcp.NewTool("post_containers_id_update",
		mcp.WithDescription("Update a container"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID or name of the container")),
		mcp.WithNumber("MemorySwappiness", mcp.Description("Input parameter: Tune a container's memory swappiness behavior. Accepts an integer between 0 and 100.")),
		mcp.WithNumber("IOMaximumIOps", mcp.Description("Input parameter: Maximum IOps for the container system drive (Windows only)")),
		mcp.WithArray("BlkioDeviceWriteBps", mcp.Description("Input parameter: Limit write rate (bytes per second) to a device, in the form `[{\"Path\": \"device_path\", \"Rate\": rate}]`.\n")),
		mcp.WithArray("Devices", mcp.Description("Input parameter: A list of devices to add to the container.")),
		mcp.WithNumber("BlkioWeight", mcp.Description("Input parameter: Block IO weight (relative weight).")),
		mcp.WithArray("Ulimits", mcp.Description("Input parameter: A list of resource limits to set in the container. For example: `{\"Name\": \"nofile\", \"Soft\": 1024, \"Hard\": 2048}`\"\n")),
		mcp.WithArray("BlkioDeviceWriteIOps", mcp.Description("Input parameter: Limit write rate (IO per second) to a device, in the form `[{\"Path\": \"device_path\", \"Rate\": rate}]`.\n")),
		mcp.WithNumber("CpuRealtimeRuntime", mcp.Description("Input parameter: The length of a CPU real-time runtime in microseconds. Set to 0 to allocate no time allocated to real-time tasks.")),
		mcp.WithArray("DeviceCgroupRules", mcp.Description("Input parameter: a list of cgroup rules to apply to the container")),
		mcp.WithNumber("CpuShares", mcp.Description("Input parameter: An integer value representing this container's relative CPU weight versus other containers.")),
		mcp.WithNumber("KernelMemory", mcp.Description("Input parameter: Kernel memory limit in bytes.")),
		mcp.WithBoolean("OomKillDisable", mcp.Description("Input parameter: Disable OOM Killer for the container.")),
		mcp.WithArray("BlkioDeviceReadBps", mcp.Description("Input parameter: Limit read rate (bytes per second) from a device, in the form `[{\"Path\": \"device_path\", \"Rate\": rate}]`.\n")),
		mcp.WithNumber("CpuPercent", mcp.Description("Input parameter: The usable percentage of the available CPUs (Windows only).\n\nOn Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.\n")),
		mcp.WithNumber("CpuQuota", mcp.Description("Input parameter: Microseconds of CPU time that the container can get in a CPU period.")),
		mcp.WithNumber("CpuPeriod", mcp.Description("Input parameter: The length of a CPU period in microseconds.")),
		mcp.WithString("CpusetMems", mcp.Description("Input parameter: Memory nodes (MEMs) in which to allow execution (0-3, 0,1). Only effective on NUMA systems.")),
		mcp.WithNumber("CpuRealtimePeriod", mcp.Description("Input parameter: The length of a CPU real-time period in microseconds. Set to 0 to allocate no time allocated to real-time tasks.")),
		mcp.WithNumber("Memory", mcp.Description("Input parameter: Memory limit in bytes.")),
		mcp.WithNumber("IOMaximumBandwidth", mcp.Description("Input parameter: Maximum IO in bytes per second for the container system drive (Windows only)")),
		mcp.WithString("CpusetCpus", mcp.Description("Input parameter: CPUs in which to allow execution (e.g., `0-3`, `0,1`)")),
		mcp.WithString("CgroupParent", mcp.Description("Input parameter: Path to `cgroups` under which the container's `cgroup` is created. If the path is not absolute, the path is considered to be relative to the `cgroups` path of the init process. Cgroups are created if they do not already exist.")),
		mcp.WithNumber("PidsLimit", mcp.Description("Input parameter: Tune a container's pids limit. Set -1 for unlimited.")),
		mcp.WithNumber("NanoCPUs", mcp.Description("Input parameter: CPU quota in units of 10<sup>-9</sup> CPUs.")),
		mcp.WithNumber("DiskQuota", mcp.Description("Input parameter: Disk limit (in bytes).")),
		mcp.WithArray("BlkioDeviceReadIOps", mcp.Description("Input parameter: Limit read rate (IO per second) from a device, in the form `[{\"Path\": \"device_path\", \"Rate\": rate}]`.\n")),
		mcp.WithArray("BlkioWeightDevice", mcp.Description("Input parameter: Block IO weight (relative device weight) in the form `[{\"Path\": \"device_path\", \"Weight\": weight}]`.\n")),
		mcp.WithNumber("MemorySwap", mcp.Description("Input parameter: Total memory limit (memory + swap). Set as `-1` to enable unlimited swap.")),
		mcp.WithNumber("MemoryReservation", mcp.Description("Input parameter: Memory soft limit in bytes.")),
		mcp.WithNumber("CpuCount", mcp.Description("Input parameter: The number of usable CPUs (Windows only).\n\nOn Windows Server containers, the processor resource controls are mutually exclusive. The order of precedence is `CPUCount` first, then `CPUShares`, and `CPUPercent` last.\n")),
		mcp.WithObject("RestartPolicy", mcp.Description("Input parameter: The behavior to apply when the container exits. The default is not to restart.\n\nAn ever increasing delay (double the previous delay, starting at 100ms) is added before each restart to prevent flooding the server.\n")),
	)

	return models.Tool{
		Definition: tool,
		Handler:    ContainerupdateHandler(cfg),
	}
}
