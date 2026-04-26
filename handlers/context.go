package handlers

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

// WithRLS wraps a database operation with RLS session variables.
// It sets 'app.current_user_id' and 'app.current_user_roles' in the transaction.
func WithRLS(c *gin.Context, dbConn *sql.DB, fn func(tx *sql.Tx) error) error {
	userID, _ := c.Get("user_id")
	userRoles, _ := c.Get("user_roles")

	tx, err := dbConn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set session variables for RLS
	// We use set_config to safely set session variables with parameters
	_, err = tx.ExecContext(c.Request.Context(), "SELECT set_config('app.current_user_id', $1, true)", fmt.Sprintf("%v", userID))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(c.Request.Context(), "SELECT set_config('app.current_user_roles', $1, true)", fmt.Sprintf("%v", userRoles))
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
