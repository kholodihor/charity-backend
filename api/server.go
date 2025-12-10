package api

import (
	"log"
	"net/http"
	"strconv"

	db "charity/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Server struct {
	router *gin.Engine
	store  *db.Store
}

func NewServer(store *db.Store) *Server {
	r := gin.Default()
	s := &Server{
		router: r,
		store:  store,
	}

	s.registerRoutes()

	return s
}

func (s *Server) Start(address string) error {
	log.Printf("starting HTTP server on %s", address)
	return s.router.Run(address)
}

func (s *Server) registerRoutes() {
	// health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	donations := s.router.Group("/donations")
	donations.POST("", s.createDonation)

	goals := s.router.Group("/goals")
	goals.GET("", s.listGoals)
	goals.GET(":id", s.getGoal)
}

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

func (s *Server) listGoals(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "list goals not implemented yet",
	})
}

func (s *Server) getGoal(c *gin.Context) {
	_ = strconv.Itoa // temporary use of strconv to avoid unused import; real implementation will parse ID

	c.JSON(http.StatusNotImplemented, gin.H{"error": "get goal not implemented yet"})
}
