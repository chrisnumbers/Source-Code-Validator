package util

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	MongoDB *mongo.Client
	RedisDB *redis.Client
}

func NewHandler(mongoClient *mongo.Client, redisClient *redis.Client) *Handler {
	return &Handler{MongoDB: mongoClient, RedisDB: redisClient}
}
