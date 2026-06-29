package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbteam "github.com/Wei-Shaw/sub2api/ent/team"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type teamRepository struct {
	client *dbent.Client
}

// NewTeamRepository creates a new TeamRepository
func NewTeamRepository(client *dbent.Client) service.TeamRepository {
	return &teamRepository{client: client}
}

func (r *teamRepository) Create(ctx context.Context, teamIn *service.Team) error {
	created, err := r.client.Team.Create().
		SetName(teamIn.Name).
		SetOwnerID(teamIn.OwnerID).
		SetInviteCode(teamIn.InviteCode).
		SetStatus(teamIn.Status).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrInviteCodeExists)
	}
	applyTeamEntityToService(teamIn, created)
	return nil
}

func (r *teamRepository) GetByID(ctx context.Context, id int64) (*service.Team, error) {
	m, err := r.client.Team.Query().Where(dbteam.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTeamNotFound, nil)
	}
	return teamEntityToService(m), nil
}

func (r *teamRepository) GetByOwnerID(ctx context.Context, ownerID int64) (*service.Team, error) {
	m, err := r.client.Team.Query().Where(dbteam.OwnerIDEQ(ownerID)).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTeamNotFound, nil)
	}
	return teamEntityToService(m), nil
}

func (r *teamRepository) GetByInviteCode(ctx context.Context, code string) (*service.Team, error) {
	m, err := r.client.Team.Query().
		Where(
			dbteam.InviteCodeEQ(code),
			dbteam.StatusEQ(service.StatusActive),
		).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTeamNotFound, nil)
	}
	return teamEntityToService(m), nil
}

func (r *teamRepository) UpdateInviteCode(ctx context.Context, teamID int64, code string) error {
	n, err := r.client.Team.Update().
		Where(dbteam.IDEQ(teamID)).
		SetInviteCode(code).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrTeamNotFound, service.ErrInviteCodeExists)
	}
	if n == 0 {
		return service.ErrTeamNotFound
	}
	return nil
}

func (r *teamRepository) Update(ctx context.Context, teamIn *service.Team) error {
	n, err := r.client.Team.Update().
		Where(dbteam.IDEQ(teamIn.ID)).
		SetName(teamIn.Name).
		SetStatus(teamIn.Status).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrTeamNotFound, nil)
	}
	if n == 0 {
		return service.ErrTeamNotFound
	}
	return nil
}

func (r *teamRepository) Delete(ctx context.Context, teamID int64) error {
	err := r.client.Team.DeleteOneID(teamID).Exec(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrTeamNotFound, nil)
	}
	return nil
}

func teamEntityToService(t *dbent.Team) *service.Team {
	if t == nil {
		return nil
	}
	return &service.Team{
		ID:         t.ID,
		Name:       t.Name,
		OwnerID:    t.OwnerID,
		InviteCode: t.InviteCode,
		Status:     t.Status,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

func applyTeamEntityToService(dst *service.Team, src *dbent.Team) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
