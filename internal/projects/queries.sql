-- name: GetProject :one
SELECT id, name, unix_name, description, active, created_at, updated_at
FROM projects
WHERE id = $1;

-- name: GetProjectByUnixName :one
SELECT id, name, unix_name, description, active, created_at, updated_at
FROM projects
WHERE unix_name = $1;

-- name: CheckProjectExistsByUnixName :one
SELECT EXISTS(
    SELECT 1 FROM projects WHERE unix_name = $1
);

-- name: CreateProject :one
INSERT INTO projects (
    id, name, unix_name, description, active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, name, unix_name, description, active, created_at, updated_at;

-- name: ListProjects :many
SELECT id, name, unix_name, description, active, created_at, updated_at
FROM projects
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(NULLIF($2, ''), name),
    description = COALESCE(sqlc.narg('description'), description),
    active = COALESCE(sqlc.narg('active'), active),
    updated_at = $3
WHERE id = $1
RETURNING id, name, unix_name, description, active, created_at, updated_at;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;
