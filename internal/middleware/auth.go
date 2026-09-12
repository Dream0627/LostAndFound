package middleware

import (
	"errors"
	//"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"LAF/config"
	"LAF/pkg/apperror"
)

const (
	UserIDKey = "auth_user_id"
	RoleKey   = "auth_role"
)

type authClaims struct {
	UserID uint64 `json:"user_id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}


func bearerToken(authHeader string) (string, error) {
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return "", errors.New("missing or malformed authorization header")
    }

    token := strings.TrimPrefix(authHeader, "Bearer ")
    if token == "" {
        return "", errors.New("token is empty")
    }
    return token, nil
}

func Auth(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil || jwtConfig.Secret == "" {
			apperror.AbortWithException(c, apperror.UnauthorizedError, err)
			return
		}

		claims := &authClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString, 
			claims, 
			func(token *jwt.Token) (any, error) {
			    if token.Method != jwt.SigningMethodHS256 { 
					return nil, errors.New("unexpected jwt signing method")
			    }
				return []byte(jwtConfig.Secret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid || claims.Role == "" || claims.UserID == 0 {
			apperror.AbortWithException(c, apperror.UnauthorizedError, err)
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		c.Next()
	}
}

func CurrentRole(c *gin.Context) (string, bool) {
	role, ok := c.Get(RoleKey)
	if !ok {
		return "", false
	}
	return role.(string), true
}

func RequireRole(allowedRoles []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        currentRole, ok := CurrentRole(c)
        
        roleAllowed := false
        for _, role := range allowedRoles {
            if currentRole == role {
                roleAllowed = true
                break
            }
        }

        if !ok || !roleAllowed {
            apperror.AbortWithError(c, apperror.AdminForbiddenError)
            return
        }
        c.Next()
    }
}

func CurrentUserID(c *gin.Context) (uint64, bool) { 
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	return userID.(uint64), true
}