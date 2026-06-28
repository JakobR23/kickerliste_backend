package httputil

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ParseID reads the "id" path parameter and writes a 400 response if it is
// not a valid integer. Returns (id, true) on success, (0, false) on failure.
func ParseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return 0, false
	}
	return id, true
}

// BindJSON binds the request body into dst, writing a 400 response on failure.
// Returns true on success; false (response already written) on failure.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return false
	}
	return true
}

// StatusError is a domain error that carries the HTTP status code it should map
// to. Domain packages declare sentinels of this type so HandleError can
// translate them centrally, without each handler needing its own errors.Is
// ladder. The Message is client-safe and is returned verbatim.
type StatusError struct {
	Status  int
	Message string
}

func (e *StatusError) Error() string { return e.Message }

// NewStatusError constructs a StatusError sentinel.
func NewStatusError(status int, message string) *StatusError {
	return &StatusError{Status: status, Message: message}
}

// HandleError maps known errors to appropriate HTTP responses.
//
//   - *StatusError                          → its Status, with its Message
//   - pgx.ErrNoRows                         → 404 Not Found
//   - PgError unique_violation (23505)      → 409 Conflict
//   - PgError foreign_key_violation (23503) → 422 Unprocessable Entity
//   - everything else                       → 500 Internal Server Error
func HandleError(c *gin.Context, err error) {
	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		c.JSON(statusErr.Status, gin.H{"message": statusErr.Message})
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"message": "resource not found"})
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			c.JSON(http.StatusConflict, gin.H{"message": pgErr.Detail})
			return
		case pgerrcode.ForeignKeyViolation:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": pgErr.Detail})
			return
		}
	}
	// Attach the full error to the Gin context so RequestLogger can persist it.
	// The client receives a generic message — internal details are never exposed.
	c.Error(err) //nolint:errcheck
	c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}
