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

func (s *Server) createDonation(c *gin.Context) {
	type request struct {
		UserID      int64  `json:"user_id"`
		GoalID      int64  `json:"goal_id"`
		Amount      int64  `json:"amount"`
		Currency    string `json:"currency"`
		IsAnonymous bool   `json:"is_anonymous"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	params := db.DonationTxParams{
		UserID: pgtype.Int8{
			Int64: req.UserID,
			Valid: true,
		},
		GoalID:      req.GoalID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		IsAnonymous: req.IsAnonymous,
	}

	result, err := s.store.DonationTx(c.Request.Context(), params)
	if err != nil {
		log.Printf("createDonation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create donation"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) getDonation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid donation id"})
		return
	}

	donation, err := s.store.GetDonation(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "donation not found"})
			return
		}
		log.Printf("getDonation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get donation"})
		return
	}

	c.JSON(http.StatusOK, donation)
}

func (s *Server) listDonationsByGoal(c *gin.Context) {
	goalIDStr := c.Param("goal_id")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil || goalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

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

	donations, err := s.store.ListDonationsByGoal(c.Request.Context(), db.ListDonationsByGoalParams{
		GoalID: goalID,
		Limit:  int32(limit64),
		Offset: int32(offset64),
	})
	if err != nil {
		log.Printf("listDonationsByGoal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list donations"})
		return
	}

	c.JSON(http.StatusOK, donations)
}

func (s *Server) listDonationsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

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

	donations, err := s.store.ListDonationsByUser(c.Request.Context(), db.ListDonationsByUserParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Limit:  int32(limit64),
		Offset: int32(offset64),
	})
	if err != nil {
		log.Printf("listDonationsByUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list donations"})
		return
	}

	c.JSON(http.StatusOK, donations)
}
