package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	configDir "github.com/alonelegion/golang-mongo-api/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server      *gin.Engine
	ctx         context.Context
	mongoClient *mongo.Client
	redisClient *redis.Client
)

func init() {
	config, err := configDir.LoadConfig("../.")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	ctx = context.TODO()

	mongoConn := options.Client().ApplyURI(config.DBUri)
	mongoClient, err = mongo.Connect(ctx, mongoConn)
	if err != nil {
		panic(err)
	}

	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	redisClient = redis.NewClient(&redis.Options{Addr: config.RedisUri})

	if _, err = redisClient.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	fmt.Println("Redis client connected successfully...")

	err = redisClient.Set(ctx, "test", "Welcome to Golang with Redis and MongoDB",
		0).Err()
	if err != nil {
		panic(err)
	}

	server = gin.Default()
}

func main() {
	config, err := configDir.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load config", err)
	}

	defer mongoClient.Disconnect(ctx)

	value, err := redisClient.Get(ctx, "test").Result()
	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		panic(err)
	}

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": value})
	})

	log.Fatal(server.Run(":" + config.Port))
}
