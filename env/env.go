package env

import (
	"bufio"
	"fmt"
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
	// LogLevel sets the minimum log level (debug, info, warn, error).
	// Defaults to "info" when not set.
	LogLevel EnvKey = "LOG_LEVEL"
	// ErrorLogFile is the path to the file where error-level log entries are
	// persisted in JSON format. Defaults to "errors.log" when not set.
	ErrorLogFile EnvKey = "ERROR_LOG_FILE"
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

// ValidateConfig checks that all required environment variables are present and
// meet minimum requirements. It writes each problem to stderr and exits with
// code 1 if any are found. Call this once at startup before using any values.
func ValidateConfig() {
	var problems []string

	for _, key := range []EnvKey{DatabaseUser, DatabasePassword, DatabaseHost, DatabasePort, DatabaseName} {
		if key.GetValue() == "" {
			problems = append(problems, fmt.Sprintf("  %s is required but not set", key))
		}
	}

	secret := JWTSecret.GetValue()
	switch {
	case secret == "":
		problems = append(problems, fmt.Sprintf("  %s is required but not set", JWTSecret))
	case len(secret) < 32:
		problems = append(problems, fmt.Sprintf("  %s must be at least 32 characters (got %d)", JWTSecret, len(secret)))
	}

	if len(problems) > 0 {
		fmt.Fprintln(os.Stderr, "startup configuration error:")
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
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
			return
		}
		fmt.Fprintf(os.Stderr, "unable to read env/.env: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "env/.env scan error: %v\n", err)
		os.Exit(1)
	}
}
