package projects

import "time"

// Project represents the core domain entity for a client project.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	UnixName    string    `json:"unix_name"`
	Description string    `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectRequest is the input payload for creating a new project.
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	UnixName    string `json:"unix_name" validate:"required,min=3,max=100,unix_name"`
	Description string `json:"description,omitempty"`
}

// UpdateProjectRequest is the payload for updating an existing project.
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}
