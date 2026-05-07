package env

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type EnvKey string

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}

const (
	DatabaseUser     EnvKey = "DB_USER"
	DatabasePassword EnvKey = "DB_PASSWORD"
	DatabaseHost     EnvKey = "DB_HOST"
	DatabasePort     EnvKey = "DB_PORT"
	Host             EnvKey = "HOST"
	Port             EnvKey = "PORT"
	HashsaltLength   EnvKey = "HASHSALT_LENGTH"
	JWTSecret        EnvKey = "JWT_SECRET"
)

func LoadConfig() {
	file, err := os.OpenFile("./env/.env", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("unable to read file %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pair := strings.SplitN(scanner.Text(), "=", 2)
		os.Setenv(pair[0], pair[1])
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
	}
}
