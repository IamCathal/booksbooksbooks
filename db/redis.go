package db

import (
	"context"
	"os"
	"time"

	redis "github.com/go-redis/redis/v9"
	"go.uber.org/zap"
)

var (
	logger      *zap.Logger
	ctx         = context.Background()
	redisClient *redis.Client

	DEFAULT_TTL = time.Duration(0)
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func ConnectToRedis() {
	logger.Info("Connecting to redis...", zap.String("diagnostics", "redis"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	redisClient = rdb
	response, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Sugar().Fatalf("Could not connect to redis. Response: '%s' error: '%s'", response, err)
	}

	logger.Info("Redis connection successfully initialised")
}
