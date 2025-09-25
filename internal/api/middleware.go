package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dottox/social/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

func (app *Application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user := app.getAuthUserFromCtx(ctx)
		post := app.getPostFromCtx(ctx)

		// If the user is the owner of the post, allow
		if post.UserId == user.Id {
			next.ServeHTTP(w, r)
			return
		}

		// If not the owner, check if they have the required role
		allowed, err := app.checkRolePrecedence(ctx, user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if !allowed {
			app.forbiddenError(w, r, fmt.Errorf("insufficient permissions to modify this resource"))
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (app *Application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// read the auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("missing Authorization header"))
			return
		}

		// parse it -> get the bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid Authorization header format"))
			return
		}

		// validate the token
		token := parts[1]
		jwtToken, err := app.Authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["sub"]), 10, 32)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		user, err := app.Store.Users.GetById(ctx, uint32(userId))
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		// add the user to the context
		ctx = context.WithValue(ctx, userAuthCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (app *Application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("missing Authorization header"))
				return
			}

			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid Authorization header format"))
				return
			}
			token := parts[1]

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(token)
			if err != nil {
				app.unauthorizedBasicError(w, r, err)
				return
			}

			// check the credentials
			username := app.Config.Auth.Basic.Username
			pass := app.Config.Auth.Basic.Password

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *Application) checkRolePrecedence(ctx context.Context, user *model.User, roleName string) (bool, error) {
	role, err := app.Store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *Application) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.RateLimiter.Enabled {
			if allow, retryAfter := app.RateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededError(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
