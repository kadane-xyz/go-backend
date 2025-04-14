package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type FriendRepository interface {
	GetFriends(ctx context.Context, params sql.GetFriendsParams) ([]sql.GetFriendsRow, error)
	GetFriendRequestStatus(ctx context.Context, params sql.GetFriendRequestStatusParams) (sql.GetFriendRequestStatusRow, error)
	CreateFriendRequest(ctx context.Context, params sql.CreateFriendRequestParams) error
	AcceptFriendRequest(ctx context.Context, params sql.AcceptFriendRequestParams) error
	BlockFriend(ctx context.Context, params sql.BlockFriendParams) error
	UnblockFriend(ctx context.Context, params sql.UnblockFriendParams) error
	DeleteFriendship(ctx context.Context, params sql.DeleteFriendshipParams) error
}

type SQLFriendRepository struct {
	queries *sql.Queries
}

func NewSQLFriendRepository(queries *sql.Queries) FriendRepository {
	return &SQLFriendRepository{queries: queries}
}

func (r *SQLFriendRepository) GetFriends(ctx context.Context, userId string) ([]sql.GetFriendsRow, error) {
	q, err := r.queries.GetFriends(ctx, userId)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLFriendRepository) GetFriendRequestStatus(ctx context.Context, params sql.GetFriendRequestStatusParams) (sql.GetFriendRequestStatusRow, error) {
	q, err := r.queries.GetFriendRequestStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *SQLFriendRepository) CreateFriendRequest(ctx context.Context, params sql.CreateFriendRequestParams) error {
	q, err := r.queries.CreateFriendRequest(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) AcceptFriendRequest(ctx context.Context, params sql.AcceptFriendRequestParams) error {
	q, err := r.queries.AcceptFriendRequest(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) BlockFriend(ctx context.Context, params sql.BlockFriendParams) error {
	q, err := r.queries.BlockFriend(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) UnblockFriend(ctx context.Context, params sql.UnblockFriendParams) error {
	q, err := r.queries.UnblockFriend(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) DeleteFriendship(ctx context.Context, params sql.DeleteFriendshipParams) error {
	q, err := r.queries.DeleteFriendship(ctx, params)
	if err != nil {
		return err
	}
	return nil
}
