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
	DatabaseName     EnvKey = "DB_NAME"
	Host             EnvKey = "HOST"
	Port             EnvKey = "PORT"
	HashsaltLength   EnvKey = "HASHSALT_LENGTH"
	JWTSecret        EnvKey = "JWT_SECRET"
	AppEnv           EnvKey = "APP_ENV"
	// AllowedOrigins is a comma-separated list of origins permitted by the CORS
	// policy, e.g. "http://localhost:3000,https://app.example.com".
	// Defaults to "http://localhost:3000" when not set.
	AllowedOrigins EnvKey = "ALLOWED_ORIGINS"
)

// GetAllowedOrigins returns the list of permitted CORS origins.
// Falls back to ["http://localhost:3000"] when ALLOWED_ORIGINS is not set.
func GetAllowedOrigins() []string {
	raw := AllowedOrigins.GetValue()
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// IsDevelopment reports whether the application is running in development mode
// (APP_ENV=development). Destructive operations such as automatic schema repair
// are only permitted when this returns true.
func IsDevelopment() bool {
	return AppEnv.GetValue() == "development"
}

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
