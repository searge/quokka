package plugin

import "context"

// Plugin represents an external system integration.
type Plugin interface {
	// Name returns the unique system name of the plugin (e.g., "proxmox").
	Name() string

	// Health checks if the plugin is correctly configured and reachable.
	Health(ctx context.Context) error

	// Provision requests the creation of external resources.
	Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResult, error)

	// Status returns the current state of a provisioned resource.
	Status(ctx context.Context, resourceID string) (*StatusResult, error)

	// Deprovision requests the destruction of external resources.
	Deprovision(ctx context.Context, resourceID string) error
}

// ProvisionRequest contains parameters for creating new external resources.
type ProvisionRequest struct {
	ProjectID   string                 `json:"project_id"`
	ProjectName string                 `json:"project_name"`
	Template    string                 `json:"template,omitempty"`
	Resources   map[string]interface{} `json:"resources,omitempty"`
}

// ProvisionResult is the result of a successful provisioning attempt.
type ProvisionResult struct {
	ResourceID string            `json:"resource_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Status     string            `json:"status"`
}

// StatusResult contains the current state of an external resource.
type StatusResult struct {
	Status   string            `json:"status"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
