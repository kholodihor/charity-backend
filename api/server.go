package api

import (
	"log"
	"net/http"
	"time"

	db "charity/db/sqlc"
	"charity/token"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router               *gin.Engine
	store                *db.Store
	tokenMaker           token.Maker
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewServer(store *db.Store, tokenMaker token.Maker, accessTokenDuration, refreshTokenDuration time.Duration) *Server {
	r := gin.Default()
	s := &Server{
		router:               r,
		store:                store,
		tokenMaker:           tokenMaker,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
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
	donations.GET(":id", s.getDonation)
	donations.GET("by_goal/:goal_id", s.listDonationsByGoal)
	donations.GET("by_user/:user_id", s.listDonationsByUser)

	goals := s.router.Group("/goals")
	goals.POST("", s.createGoal)
	goals.GET("", s.listGoals)
	goals.GET(":id", s.getGoal)
	goals.PATCH(":id", s.updateGoal)

	users := s.router.Group("/users")
	users.POST("", s.createUser)
	users.POST("/login", s.loginUser)
	users.GET("", s.listUsers)
	users.GET(":id", s.getUser)
	users.GET("/by-email", s.getUserByEmail)
}
