package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type FriendRepository interface {
	GetFriends(ctx context.Context, userId string) ([]domain.Friend, error)
	GetFriendRequestStatus(ctx context.Context, params sql.GetFriendRequestStatusParams) (domain.FriendRequestStatus, error)
	CreateFriendRequest(ctx context.Context, params sql.CreateFriendRequestParams) error
	AcceptFriendRequest(ctx context.Context, params sql.AcceptFriendRequestParams) error
	BlockFriend(ctx context.Context, params sql.BlockFriendParams) error
	UnblockFriend(ctx context.Context, params sql.UnblockFriendParams) error
	DeleteFriendship(ctx context.Context, params sql.DeleteFriendshipParams) error
	GetFriendRequestsSent(ctx context.Context, userId string) ([]domain.FriendRequest, error)
	GetFriendByUsername(ctx context.Context, username string) ([]domain.Friend, error)
}

type SQLFriendRepository struct {
	queries *sql.Queries
}

func NewSQLFriendRepository(queries *sql.Queries) *SQLFriendRepository {
	return &SQLFriendRepository{queries: queries}
}

func (r *SQLFriendRepository) GetFriends(ctx context.Context, userId string) ([]sql.GetFriendsRow, error) {
	q, err := r.queries.GetFriends(ctx, userId)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLFriendRepository) GetFriendRequestStatus(ctx context.Context, params sql.GetFriendRequestStatusParams) (domain.FriendRequestStatus, error) {
	q, err := r.queries.GetFriendRequestStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLFriendRepository) CreateFriendRequest(ctx context.Context, params sql.CreateFriendRequestParams) error {
	err := r.queries.CreateFriendRequest(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) AcceptFriendRequest(ctx context.Context, params sql.AcceptFriendRequestParams) error {
	err := r.queries.AcceptFriendRequest(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) BlockFriend(ctx context.Context, params sql.BlockFriendParams) error {
	err := r.queries.BlockFriend(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) UnblockFriend(ctx context.Context, params sql.UnblockFriendParams) error {
	err := r.queries.UnblockFriend(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) DeleteFriendship(ctx context.Context, params sql.DeleteFriendshipParams) error {
	err := r.queries.DeleteFriendship(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) GetFriendRequestsSent(ctx context.Context, userId string) ([]sql.GetFriendRequestsSentRow, error) {
	q, err := r.queries.GetFriendRequestsSent(ctx, userId)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLFriendRepository) GetFriendsByUsername(ctx context.Context, username string) ([]domain.Friend, error) {
	q, err := r.queries.GetFriendsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return domain.FromSQLFriendsByUsernameRows(q), nil
}
