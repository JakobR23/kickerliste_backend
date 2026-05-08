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

// HandleError maps common database errors to appropriate HTTP responses.
//
//   - pgx.ErrNoRows                       → 404 Not Found
//   - PgError unique_violation (23505)    → 409 Conflict
//   - PgError foreign_key_violation (23503) → 422 Unprocessable Entity
//   - everything else                     → 500 Internal Server Error
func HandleError(c *gin.Context, err error) {
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
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}
