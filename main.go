package charity

import (
	"context"
	"log"
	"time"

	"charity/api"
	"charity/config"
	db "charity/db/sqlc"
	"charity/token"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer conn.Close(context.Background())

	store := db.NewStore(conn)

	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		log.Fatalf("cannot create token maker: %v", err)
	}

	server := api.NewServer(store, tokenMaker, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

	if err := server.Start(cfg.ServerAddress); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}
