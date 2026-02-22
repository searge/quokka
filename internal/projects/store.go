package projects

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/searge/quokka/internal/projects/db"
)

// Store provides data access for project entities via sqlc.
type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewStore initializes a new Store instance.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: db.New(pool),
	}
}

// Create inserts a new project.
func (s *Store) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	id := uuid.New()

	desc := pgtype.Text{}
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	params := db.CreateProjectParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        req.Name,
		UnixName:    req.UnixName,
		Description: desc,
		Active:      true,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	row, err := s.queries.CreateProject(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrProjectExists
		}
		return nil, err
	}

	return mapToDomainProject(row), nil
}

// GetByID retrieves a project by its unique ID.
func (s *Store) GetByID(ctx context.Context, id string) (*Project, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidProjectID
	}

	row, err := s.queries.GetProject(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return nil, err
	}

	return mapToDomainProject(row), nil
}

// GetByUnixName retrieves a project by its unix name.
func (s *Store) GetByUnixName(ctx context.Context, unixName string) (*Project, error) {
	row, err := s.queries.GetProjectByUnixName(ctx, unixName)
	if err != nil {
		return nil, err
	}
	return mapToDomainProject(row), nil
}

// ExistsByUnixName checks if a project unix name is already taken.
func (s *Store) ExistsByUnixName(ctx context.Context, unixName string) (bool, error) {
	return s.queries.CheckProjectExistsByUnixName(ctx, unixName)
}

// List retrieves a list of active projects securely.
func (s *Store) List(ctx context.Context, limit, offset int32) ([]*Project, error) {
	rows, err := s.queries.ListProjects(ctx, db.ListProjectsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, len(rows))
	for i, row := range rows {
		projects[i] = mapToDomainProject(row)
	}

	return projects, nil
}

// Update amends the details of an existing project.
func (s *Store) Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidProjectID
	}

	params := db.UpdateProjectParams{
		ID:        pgtype.UUID{Bytes: uid, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	if req.Name != nil {
		params.Column2 = *req.Name
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.Active != nil {
		params.Active = pgtype.Bool{Bool: *req.Active, Valid: true}
	}

	row, err := s.queries.UpdateProject(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToDomainProject(row), nil
}

// Delete removes a project permanently.
func (s *Store) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidProjectID
	}
	rowsAffected, err := s.queries.DeleteProject(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func mapToDomainProject(row db.Project) *Project {
	return &Project{
		ID:          uuid.UUID(row.ID.Bytes).String(),
		Name:        row.Name,
		UnixName:    row.UnixName,
		Description: row.Description.String,
		Active:      row.Active,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
