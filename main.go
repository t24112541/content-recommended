package main

import (
	"content-recommended/config"
	dbPostgres "content-recommended/database/postgres"
	dbRedis "content-recommended/database/redis"
	"content-recommended/router"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load environment: %s", color.RedString(err.Error()))
	}

	HttpServer := config.NewServer()
	app, err := HttpServer.RunServer()
	if err != nil {
		strErr := fmt.Sprintf("Failed to start server: %s", color.RedString(err.Error()))
		log.Fatal("Failed to RunServer: ", strErr)
	}

	// go routines for connect DB
	dbChan := make(chan *gorm.DB, 1)
	go func() {
		db, err := dbPostgres.NewConnection()
		if err != nil {
			log.Fatalf("Failed to connect to database: %s", color.RedString(err.Error()))
		}

		gormDB, dbErr := db.Connect()
		if dbErr != nil {
			log.Fatalf("Error database connection: %s", color.RedString(dbErr.Error()))
		}
		dbChan <- gormDB
	}()
	db := <-dbChan

	redisChan := make(chan *redis.Client, 1)
	go func() {
		redisConf, err := dbRedis.NewRedisConnection()
		if err != nil {
			log.Fatalf("Failed to parse redis config: %s", color.RedString(err.Error()))
		}

		redisClient, redisErr := redisConf.Connect()
		if redisErr != nil {
			log.Fatalf("Failed to connect to redis: %s", color.RedString(redisErr.Error()))
		}
		redisChan <- redisClient
	}()
	redisClient := <-redisChan

	sessionStore := session.New(session.Config{
		Expiration: 5 * time.Minute,
	})
	validator := validator.New()

	app = config.NewConfigRouter(
		db,
		validator,
	).Config(app)

	router.New(
		sessionStore,
		db,
		redisClient,
		validator,
	).Init(app)

	// Start the HTTP server in a goroutine
	go func() {
		if err := app.Listen(":" + HttpServer.AppPort); err != nil {
			strErr := fmt.Sprintf("Failed to start server: %s", color.RedString(err.Error()))
			log.Fatal(strErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Gracefully shutting down...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("Server forced to shutdown: %s", color.RedString(err.Error()))
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %s", color.RedString(err.Error()))
		}
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing redis: %s", color.RedString(err.Error()))
	}

	log.Println("Server exited successfully.")
}
