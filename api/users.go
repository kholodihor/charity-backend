package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	db "charity/db/sqlc"
	"charity/token"
	"charity/util"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	var namePtr *string
	if user.Name.Valid {
		name := user.Name.String
		namePtr = &name
	}

	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      namePtr,
		CreatedAt: user.CreatedAt,
	}
}

func (s *Server) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := validateCreateUserRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("createUser hash error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	params := db.CreateUserParams{
		Email: req.Email,
		Name: pgtype.Text{
			Valid: req.Name != nil,
		},
		Password: pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		},
	}
	if req.Name != nil {
		params.Name.String = *req.Name
	}

	user, err := s.store.CreateUser(c.Request.Context(), params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		log.Printf("createUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

func (s *Server) getUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := s.store.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("getUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

func (s *Server) getUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
		return
	}

	user, err := s.store.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("getUserByEmail error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

func (s *Server) listUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit64, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit64 <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}
	offset64, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset64 < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return
	}

	users, err := s.store.ListUsers(c.Request.Context(), db.ListUsersParams{
		Limit:  int32(limit64),
		Offset: int32(offset64),
	})
	if err != nil {
		log.Printf("listUsers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	responses := make([]userResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, newUserResponse(user))
	}

	c.JSON(http.StatusOK, responses)
}

func (s *Server) loginUser(c *gin.Context) {
	var req loginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := validateLoginUserRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		log.Printf("loginUser get user error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	if !user.Password.Valid {
		log.Printf("loginUser missing password for user %d", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	if err := util.CheckPassword(req.Password, user.Password.String); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	const userRole = "user"
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Email, userRole, s.accessTokenDuration, token.TokenTypeAccessToken)
	if err != nil {
		log.Printf("loginUser create access token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Email, userRole, s.refreshTokenDuration, token.TokenTypeRefreshToken)
	if err != nil {
		log.Printf("loginUser create refresh token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":                     newUserResponse(user),
		"access_token":             accessToken,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token":            refreshToken,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
	})
}
