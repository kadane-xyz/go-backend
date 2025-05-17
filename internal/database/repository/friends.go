package repository

import (
	"context"

	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
)

type FriendRepository interface {
	GetFriends(ctx context.Context, userId string) ([]*domain.Friend, error)
	GetFriendRequestStatus(ctx context.Context, params *domain.FriendRequesStatusParams) (*domain.FriendshipStatus, error)
	CreateFriendRequest(ctx context.Context, params *domain.FriendRequestCreateParams) error
	AcceptFriendRequest(ctx context.Context, params *domain.FriendRequestAcceptParams) error
	BlockFriend(ctx context.Context, params *domain.FriendBlockParams) error
	UnblockFriend(ctx context.Context, params *domain.FriendUnblockParams) error
	DeleteFriendship(ctx context.Context, params *domain.FriendshipDeleteParams) error
	GetFriendRequestsSent(ctx context.Context, userId string) ([]*domain.FriendRequest, error)
	GetFriendRequestReceived(ctx context.Context, userId string) ([]*domain.FriendRequest, error)
	GetFriendByUsername(ctx context.Context, username string) ([]*domain.Friend, error)
}

type SQLFriendRepository struct {
	queries *sql.Queries
}

func NewSQLFriendRepository(queries *sql.Queries) *SQLFriendRepository {
	return &SQLFriendRepository{queries: queries}
}

func (r *SQLFriendRepository) GetFriends(ctx context.Context, userId string) ([]*domain.Friend, error) {
	q, err := r.queries.GetFriends(ctx, userId)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLFromSQLGetFriendsRow(q), nil
}

func (r *SQLFriendRepository) GetFriendRequestStatus(ctx context.Context, params *domain.FriendRequesStatusParams) (*domain.FriendshipStatus, error) {
	q, err := r.queries.GetFriendRequestStatus(ctx, sql.GetFriendRequestStatusParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *SQLFriendRepository) CreateFriendRequest(ctx context.Context, params *domain.FriendRequestCreateParams) error {
	err := r.queries.CreateFriendRequest(ctx, sql.CreateFriendRequestParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) AcceptFriendRequest(ctx context.Context, params *domain.FriendRequestAcceptParams) error {
	err := r.queries.AcceptFriendRequest(ctx, sql.AcceptFriendRequestParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) BlockFriend(ctx context.Context, params *domain.FriendBlockParams) error {
	err := r.queries.BlockFriend(ctx, sql.BlockFriendParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) UnblockFriend(ctx context.Context, params *domain.FriendUnblockParams) error {
	err := r.queries.UnblockFriend(ctx, sql.UnblockFriendParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) DeleteFriendship(ctx context.Context, params *domain.FriendshipDeleteParams) error {
	err := r.queries.DeleteFriendship(ctx, sql.DeleteFriendshipParams{
		FriendName: params.FriendName,
		UserID:     params.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLFriendRepository) GetFriendRequestsSent(ctx context.Context, userId string) ([]*domain.FriendRequest, error) {
	q, err := r.queries.GetFriendRequestsSent(ctx, userId)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetFriendRequestsSentRows(q), nil
}

func (r *SQLFriendRepository) GetFriendRequestReceived(ctx context.Context, userId string) ([]*domain.FriendRequest, error) {
	q, err := r.queries.GetFriendRequestsReceived(ctx, userId)
	if err != nil {
		return nil, err
	}
	return domain.FromSQLGetFriendRequestsReceivedRows(q), nil
}

func (r *SQLFriendRepository) GetFriendByUsername(ctx context.Context, username string) ([]*domain.Friend, error) {
	q, err := r.queries.GetFriendsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return domain.FromSQLFriendsByUsernameRows(q), nil
}
