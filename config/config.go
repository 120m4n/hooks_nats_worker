package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	NatsURL             string
	MongoURI            string
	DatabaseName        string
	Coor_CollectionName string
	Hook_CollectionName string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default values")
	}

	return Config{
		NatsURL:             getEnv("NATS_URL", "nats://localhost:4222"),
		MongoURI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:        getEnv("DATABASE_NAME", "test"),
		Coor_CollectionName: getEnv("COORDINATE_COLLECTION_NAME", "coordinates"),
		Hook_CollectionName: getEnv("HOOK_COLLECTION_NAME", "hooks"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
