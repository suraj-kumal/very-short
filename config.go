package main

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	PORT        string
	SITE_URL     string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		PORT:        ":" + os.Getenv("PORT"),
		SITE_URL:     os.Getenv("SITE_URL"),
	}
}
