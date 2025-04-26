package context

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const cookieName = "user_id"
const varDelimiter = "\r"

type Context struct {
	base          context.Context
	UserID        string
	PathVariables map[string]string
}

func From(w http.ResponseWriter, r *http.Request) *Context {
	ctx := &Context{
		base: r.Context(),
	}

	getIdFromCookie(ctx, r, w)
	getPathVariablesFromRequest(ctx, r)

	return ctx
}

func getIdFromCookie(ctx *Context, r *http.Request, w http.ResponseWriter) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		// No valid user id present; generate a new one.
		newUUID := uuid.New().String()
		cookie = &http.Cookie{
			Name:     cookieName,
			Value:    newUUID,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: false,
		}
		ctx.UserID = newUUID
	} else {
		// Extend the lifetime of the existing cookie.
		cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
		ctx.UserID = cookie.Value
	}
	// Set the cookie in the response.
	http.SetCookie(w, cookie)
}

func getPathVariablesFromRequest(ctx *Context, r *http.Request) {
	ctx.PathVariables = make(map[string]string)
	for key, values := range r.URL.Query() {
		switch len(values) {
		case 0:
			continue
		case 1:
			ctx.PathVariables[key] = values[0]
		default:
			ctx.PathVariables[key] = strings.Join(values, varDelimiter)
		}
	}
}

func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.base.Deadline()
}

func (ctx *Context) Done() <-chan struct{} {
	return ctx.base.Done()
}

func (ctx *Context) Err() error {
	return ctx.base.Err()
}

const (
	UserID = "UserID"
)

func (ctx *Context) Value(key any) any {
	if stringKey, ok := key.(string); ok {
		if stringKey == UserID {
			return ctx.UserID
		}
		if value, ok := ctx.PathVariables[stringKey]; ok {
			return value
		}
	}
	return ctx.base.Value(key)
}
