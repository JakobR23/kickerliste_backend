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

// LoadConfig reads key=value pairs from ./env/.env and sets any that are not
// already present in the environment. Variables already set (e.g. injected by
// Docker Compose) take precedence over file values, so this is safe to call
// in both local and containerised environments.
//
// If the file does not exist the function logs a warning and returns — the
// application can still start as long as all required variables were supplied
// through the environment directly.
func LoadConfig() {
	file, err := os.OpenFile("./env/.env", os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("env/.env not found — relying on environment variables")
			return
		}
		log.Fatalf("unable to read env/.env: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		// Only set the variable if it has not already been provided by the
		// environment (e.g. via Docker Compose or a shell export).
		if _, already := os.LookupEnv(key); !already {
			os.Setenv(key, pair[1])
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("env/.env scan error: %v", err)
	}
}
