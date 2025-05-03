package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alpes214/stellar-hooks/internal/api"
	"github.com/alpes214/stellar-hooks/internal/config"
	"github.com/alpes214/stellar-hooks/internal/horizon"
	"github.com/alpes214/stellar-hooks/internal/storage"
	"github.com/alpes214/stellar-hooks/internal/stream/jetstream"
	"github.com/gin-gonic/gin"

	_ "github.com/alpes214/stellar-hooks/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.LoadEnv()

	db := storage.InitPostgres()
	defer db.Close()

	store := storage.NewPostgresStore(db)
	if err := storage.MigratePostgres(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := jetstream.Connect(); err != nil {
		log.Fatalf("JetStream connection failed: %v", err)
	}
	defer jetstream.NatsConn.Drain()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := jetstream.NewJetStreamConsumer("stellar.events", "webhook-subscriber", store)
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start JetStream consumer: %v", err)
	}

	go horizon.StartSSEListenerJetStream()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := gin.Default()
	api.RegisterRoutes(r, store)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutting down gracefully...")
}
