package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string
	PORT                string
	SiteURL             string
	MixMultiplierSecret int
}

func Load() Config {
	_ = godotenv.Load()
	//if err != nil {
	//panic("Error loading .env file")
	//}
	multiplier, err := strconv.Atoi(os.Getenv("MIX_MULTIPLIER_SECRET"))
	if err != nil {
		panic("MIX_MULTIPLIER_SECRET must be a number")
	}

	return Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		PORT:                ":" + os.Getenv("PORT"),
		SiteURL:             os.Getenv("SITE_URL"),
		MixMultiplierSecret: multiplier,
	}
}
