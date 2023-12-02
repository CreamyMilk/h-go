package utils

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func LoadEnvironmentVariables() {
	env := os.Getenv("APP_ENV")

	switch env {
	case "development":
		log.Info().Msgf("Loading configurations...Development")
		err := godotenv.Load(".env.development.local")
		if err != nil {
			log.Fatal().Msg("Error loading .env.development.local file")
		}
	case "test":
		log.Print("Loading configurations...Test")
		err := godotenv.Load(".env.test.local")
		if err != nil {
			log.Fatal().Msg("Error loading .env.test.local file")
		}
	default:
		log.Print("Loading Production configuration")
		err := godotenv.Load()
		if err != nil {
			log.Fatal().Msg("Error loading .env file")
		}
	}
}

func IsProduction() (bool, error) {
	appEnv := os.Getenv("APP_ENV")
	switch appEnv {
	case "production":
		return true, nil
	case "staging":
		return false, nil
	case "development":
		return false, nil
	case "test":
		return false, nil
	default:
		return false, fmt.Errorf("unknown environment %s", appEnv)
	}
}

func IsStagingOrDev() (bool, error) {
	appEnv := os.Getenv("APP_ENV")
	switch appEnv {
	case "production":
		return false, nil
	case "staging":
		return true, nil
	case "development":
		return true, nil
	case "test":
		return false, nil
	default:
		return false, fmt.Errorf("unknown environment %s", appEnv)
	}
}

func UseEnvOrDefault(envKey, _default string) string {
	var env string
	if os.Getenv(envKey) == "" {
		env = _default
	} else {
		env = os.Getenv(envKey)
	}

	return env
}

func CheckEnvInclusion(envKeys ...string) bool {
	for _, envKey := range envKeys {
		if os.Getenv(envKey) == "" {
			return false
		}
	}
	return true
}