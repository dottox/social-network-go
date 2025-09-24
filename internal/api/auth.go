package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dottox/social/internal/mailer"
	"github.com/dottox/social/internal/model"
	"github.com/dottox/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// @Summary		Register a new user
// @Description	Register a new user with the given information
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		model.RegisterUserPayload	true	"User credentials"
// @Success		201		{object}	model.UserWithToken			"User registered successfully"
// @Failure		400		{object}	error
// @Failure		409		{object}	error
// @Failure		500		{object}	error
// @Router			/auth/user [post]
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Took the payload from the request body
	var payload model.RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Validate the struct of the payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Create the new user if the payload had no errors
	user := &model.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// Hash the user password and set the password to the user
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Generate a token for email verification
	plainToken := uuid.New().String()

	// Hash the token to store it securely in the database
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// Create the new user in the repository
	if err := app.Store.Users.CreateAndInvite(ctx, user, hashToken, app.Config.Mail.Exp); err != nil {
		switch {
		case errors.Is(err, store.ErrUsersDuplicateEmail) || errors.Is(err, store.ErrUsersDuplicateUsername):
			app.resourceAlreadyExists(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	// Returning the user with the plain token for verification
	userWithToken := &model.UserWithToken{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/activate?token=%s", app.Config.FrontendURL, plainToken)

	isProdEnv := app.Config.Env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// Send the welcome email with the activation token
	err := app.Mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.Logger.Errorw("error sending welcome email", "error", err, "user_id", user.Id, "email", user.Email)

		if err := app.Store.Users.DeleteById(ctx, user.Id); err != nil {
			app.Logger.Errorw("error deleting user", "user_id", user.Id, "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// @Summary		Activate a user account
// @Description	Activate a user account using the provided token
// @Tags			auth
// @Param			token	query	string	true	"Activation token"
// @Success		204
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Failure		500	{object}	error
// @Router			/auth/user/activate [put]
func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.badRequestError(w, r, errors.New("token is required"))
	}

	// Hash the token
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	err := app.Store.Users.Activate(r.Context(), hashToken)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrResourceNotFound):
			app.resourceNotFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// @Summary		Generate a JWT token for a user
// @Description	Generate a JWT token for a user using their email and password
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			credentials	body		CreateUserTokenPayload	true	"User credentials"
// @Success		201			{object}	string					"JWT token"
// @Failure		400			{object}	error
// @Failure		401			{object}	error
// @Failure		500			{object}	error
// @Router			/auth/token [post]
func (app *Application) getTokenHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	// parse payload credentials
	payload := CreateUserTokenPayload{}
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// fetch the user (check if the user exists) from the payload
	user, err := app.Store.Users.GetByEmail(ctx, payload.Email)
	if err != nil {
		switch err {
		case store.ErrResourceNotFound:
			app.unauthorizedError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	ok := user.Password.Matches(payload.Password)
	if !ok {
		app.unauthorizedError(w, r, errors.New("invalid credentials"))
		return
	}

	// generate the token -> add claims
	claims := jwt.MapClaims{
		"sub": user.Id,                                          // Subject
		"exp": time.Now().Add(app.Config.Auth.Token.Exp).Unix(), // Expiration time
		"iat": time.Now().Unix(),                                // Issued at
		"nbf": time.Now().Unix(),                                // Not before
		"iss": app.Config.Auth.Token.Iss,                        // Issuer
		"aud": app.Config.Auth.Token.Iss,                        // Audience
	}
	token, err := app.Authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// return the token to the user
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
