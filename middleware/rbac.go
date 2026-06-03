package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/gin-gonic/gin"
)

func NewCasbinEnforcer(dbConn *sql.DB) (*casbin.Enforcer, error) {
	a := fileadapter.NewAdapter("rbac_policy.csv")

	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (keyMatch(r.obj, p.obj) || keyMatch2(r.obj, p.obj) || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
`)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return enforcer, nil
}

func RBACMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userRoles, exists := c.Get("user_roles")
		if !exists {
			fmt.Printf("[RBAC DEBUG] Access denied for user %v: missing user roles in context\n", userID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: missing user roles"})
			return
		}

		roles := strings.Split(userRoles.(string), ",")
		
		// ALWAYS check FullPath first for parametrized routes
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		fmt.Printf("[RBAC DEBUG] User %v with roles [%s] requesting %s %s\n", userID, userRoles, method, path)

		for _, role := range roles {
			r := strings.TrimSpace(role)
			
			// 1. Try matching with the path provided by Gin (which is the pattern /api/v1/categories/:id)
			allowed, err := enforcer.Enforce(r, path, method)
			if err == nil && allowed {
				fmt.Printf("[RBAC DEBUG] Allowed via Pattern Match: role=%s, path=%s\n", r, path)
				c.Next()
				return
			}

			// 2. Try matching with raw URL path (concrete /api/v1/categories/123)
			rawPath := c.Request.URL.Path
			allowedRaw, errRaw := enforcer.Enforce(r, rawPath, method)
			if errRaw == nil && allowedRaw {
				fmt.Printf("[RBAC DEBUG] Allowed via RawPath Match: role=%s, path=%s\n", r, rawPath)
				c.Next()
				return
			}
		}

		fmt.Printf("[RBAC DEBUG] Final Result: Forbidden for user %v\n", userID)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	}
}

// Global Middleware for debugging
func RBACDebugMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
    }
}
