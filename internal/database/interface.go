package database

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserFromUsername(ctx context.Context, username string) (User, error)
	GetUsernameFromID(ctx context.Context, id uuid.UUID) (string, error)
	ChangePasswordAndUsername(ctx context.Context, arg ChangePasswordAndUsernameParams) (User, error)
}
