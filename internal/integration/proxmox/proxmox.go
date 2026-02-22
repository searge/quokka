package proxmox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/searge/quokka/internal/plugin"
)

// Plugin implements the plugin.Plugin interface for Proxmox via forge-ovh-cli.
type Plugin struct {
	cliPath string
}

// New creates a new Proxmox plugin instance.
func New(cliPath string) *Plugin {
	if cliPath == "" {
		cliPath = "forge-ovh-cli"
	}
	return &Plugin{cliPath: cliPath}
}

// Name returns the identifier for this plugin.
func (p *Plugin) Name() string {
	return "proxmox"
}

// Health verifies that the CLI is executable.
func (p *Plugin) Health(ctx context.Context) error {
	// Simple check: can we find the executable and run a help or version command?
	_, err := exec.LookPath(p.cliPath)
	if err != nil {
		return fmt.Errorf("forge-ovh-cli not found in path: %w", err)
	}

	cmd := exec.CommandContext(ctx, p.cliPath, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute forge-ovh-cli: %w", err)
	}
	return nil
}

// Provision invokes the CLI to create a new VM/container for the project.
func (p *Plugin) Provision(ctx context.Context, req plugin.ProvisionRequest) (*plugin.ProvisionResult, error) {
	// Assuming forge-ovh-cli usage: forge-ovh-cli create --name <project_name>
	// The implementation here depends on the exact CLI expected format.

	args := []string{"create", "--name", req.ProjectName}
	if req.Template != "" {
		args = append(args, "--template", req.Template)
	}

	cmd := exec.CommandContext(ctx, p.cliPath, args...)

	// Optional: pass down environment variables if CLI relies on them for auth
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	outStr := string(output)
	if err != nil {
		return nil, fmt.Errorf("forge-ovh-cli provision failed: %w, output: %s", err, outStr)
	}

	// Pseudo-parsing to get a resource ID and status
	// In reality we would parse the JSON output of forge-ovh-cli.
	resourceID := p.parseResourceID(outStr)
	if resourceID == "" {
		return nil, fmt.Errorf("unable to parse resource id from cli output")
	}

	return &plugin.ProvisionResult{
		ResourceID: resourceID,
		Status:     "provisioned",
		Metadata: map[string]string{
			"cli_output": outStr,
			"node":       "proxmox-01", // stub
		},
	}, nil
}

// Status checks the status of an existing resource via the CLI.
func (p *Plugin) Status(ctx context.Context, resourceID string) (*plugin.StatusResult, error) {
	cmd := exec.CommandContext(ctx, p.cliPath, "status", "--id", resourceID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w, output: %s", err, string(output))
	}

	// Stub parsing
	return &plugin.StatusResult{
		Status: "running",
		Metadata: map[string]string{
			"raw_output": string(output),
		},
	}, nil
}

// Deprovision removes the resource.
func (p *Plugin) Deprovision(ctx context.Context, resourceID string) error {
	cmd := exec.CommandContext(ctx, p.cliPath, "delete", "--id", resourceID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w, output: %s", err, string(output))
	}
	return nil
}

// parseResourceID is a helper to extract a resource ID from CLI output.
func (p *Plugin) parseResourceID(output string) string {
	// A naive extraction. If the CLI outputs JSON, this should be json.Unmarshal.
	lines := strings.Split(output, "\n")
	for _, l := range lines {
		if strings.HasPrefix(strings.ToLower(l), "id:") {
			return strings.TrimSpace(strings.SplitN(l, ":", 2)[1])
		}
	}
	// Fallback to generating a pseudo ID if not found for spike purpose
	return ""
}
