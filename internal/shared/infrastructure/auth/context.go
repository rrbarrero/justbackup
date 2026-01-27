package auth

import (
	"context"
)

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}
