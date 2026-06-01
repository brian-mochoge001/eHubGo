package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/gin-gonic/gin"
)

func NewCasbinEnforcer(dbConn *sql.DB) (*casbin.Enforcer, error) {
	// Fallback to FileAdapter to resolve persistence issues
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
m = g(r.sub, p.sub) && (keyMatch2(r.obj, p.obj) || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
`)
	if err != nil {
		return nil, err
	}

	// Create Enforcer with model and adapter explicitly
	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	// Load policy
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return enforcer, nil
}

func RBACMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: missing user roles"})
			return
		}

		roles := strings.Split(userRoles.(string), ",")
		for _, role := range roles {
			allowed, err := enforcer.Enforce(strings.TrimSpace(role), c.FullPath(), c.Request.Method)
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	}
}
