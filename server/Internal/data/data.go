package data

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DataStorage struct {
	Redis_db *redis.Client
	Db_pool  *pgxpool.Pool
	schema   string
}

func NewDataStorage(ctx context.Context, dbURL string, redisURL string, schema string) (*DataStorage, error) {
	//load DB
	log.Printf("Attempting to connect to database: %s", dbURL)
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	//load redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	log.Printf("Attempting to connect to redis: %s", redisURL)
	redis_client := redis.NewClient(opt)

	return &DataStorage{Redis_db: redis_client, Db_pool: pool, schema: schema}, nil
}
