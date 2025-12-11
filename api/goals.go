package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	db "charity/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) listGoals(c *gin.Context) {
	// Optional query params: limit, offset, active
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	activeStr := c.DefaultQuery("active", "")

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

	ctx := c.Request.Context()
	var goals []db.Goal
	if activeStr == "true" {
		goals, err = s.store.ListActiveGoals(ctx, db.ListActiveGoalsParams{
			Limit:  int32(limit64),
			Offset: int32(offset64),
		})
	} else {
		goals, err = s.store.ListGoals(ctx, db.ListGoalsParams{
			Limit:  int32(limit64),
			Offset: int32(offset64),
		})
	}
	if err != nil {
		log.Printf("listGoals error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list goals"})
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (s *Server) getGoal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	goal, err := s.store.GetGoal(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
			return
		}
		log.Printf("getGoal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get goal"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (s *Server) createGoal(c *gin.Context) {
	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := validateCreateGoalRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := db.CreateGoalParams{
		Title: req.Title,
		Description: pgtype.Text{
			Valid: req.Description != nil,
		},
		TargetAmount: pgtype.Int8{
			Int64: req.TargetAmount,
			Valid: true,
		},
	}
	if req.Description != nil {
		params.Description.String = *req.Description
	}

	goal, err := s.store.CreateGoal(c.Request.Context(), params)
	if err != nil {
		log.Printf("createGoal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create goal"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (s *Server) updateGoal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	var req updateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := validateUpdateGoalRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := db.UpdateGoalParams{ID: id}
	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.TargetAmount != nil {
		if *req.TargetAmount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_amount must be positive"})
			return
		}
		params.TargetAmount = pgtype.Int8{Int64: *req.TargetAmount, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}

	goal, err := s.store.UpdateGoal(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
			return
		}
		log.Printf("updateGoal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update goal"})
		return
	}

	c.JSON(http.StatusOK, goal)
}
