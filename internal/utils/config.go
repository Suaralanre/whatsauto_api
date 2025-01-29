package utils

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var logger = CustomLogger()

func init() {
	if err := godotenv.Load(); err != nil {
		logger.Error(".env files does not exist..")
	}
}

// ValidateEnv checks if a string environment variable exists
func GetEnv(envName, defaultVal string) string {
	envar, exists := os.LookupEnv(envName)
	if !exists {
		return defaultVal
	}
	return envar
}

// ValidateEnv checks if an Integer environment variable exists
func GetEnvInt(envName string, defaultVal int) int {
	valueStr := GetEnv(envName, "")

	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func GetEnvBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

func GetEnvSlice(name string, defaultVal []string, sep string) []string {
	valStr := GetEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
