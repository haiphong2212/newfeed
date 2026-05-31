package platform

import (
	"fmt"

	"github.com/newfeed/community-news/services/shared/env"
)

type Config struct {
	ServiceName      string
	Env              string
	HTTPAddr         string
	GRPCAddr         string
	PostgresDSN      string
	RedisAddr        string
	RabbitMQURL      string
	ElasticsearchURL string
	ObjectRoot       string
}

func Load(serviceName, defaultHTTP, defaultGRPC string) Config {
	dbHost := env.String("DB_HOST", "localhost")
	dbPort := env.String("DB_PORT", "5432")
	dbUser := env.String("DB_USER", "postgres")
	dbPass := env.String("DB_PASSWORD", "postgres")
	dbName := env.String("DB_NAME", "newfeed")
	rabbitUser := env.String("RABBITMQ_USER", "guest")
	rabbitPass := env.String("RABBITMQ_PASSWORD", "guest")
	rabbitHost := env.String("RABBITMQ_HOST", "localhost")
	rabbitPort := env.String("RABBITMQ_PORT", "5672")

	return Config{
		ServiceName:      serviceName,
		Env:              env.String("ENV", "development"),
		HTTPAddr:         env.String("HTTP_ADDR", defaultHTTP),
		GRPCAddr:         env.String("GRPC_ADDR", defaultGRPC),
		PostgresDSN:      env.String("POSTGRES_DSN", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)),
		RedisAddr:        env.String("REDIS_ADDR", env.String("REDIS_HOST", "localhost")+":"+env.String("REDIS_PORT", "6379")),
		RabbitMQURL:      env.String("RABBITMQ_URL", fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitHost, rabbitPort)),
		ElasticsearchURL: env.String("ELASTICSEARCH_URL", "http://localhost:9200"),
		ObjectRoot:       env.String("OBJECT_STORAGE_ROOT", "./data/object-storage"),
	}
}
