package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type SessionContext struct {
	UserID uuid.UUID
	Token  string
}

const contextKey = "user_session_context"

func CreateSessionContext(ctx context.Context, sessionCtx SessionContext) context.Context {
	return context.WithValue(ctx, contextKey, sessionCtx)
}

func GetSessionContext(r *http.Request) (SessionContext, error) {
	sessionCtx, ok := r.Context().Value(contextKey).(SessionContext)
	if !ok {
		return SessionContext{}, errors.New("session err: failed to retrieve user session context")
	}

	return sessionCtx, nil
}
