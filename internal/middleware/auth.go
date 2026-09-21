// Package middleware 提供 HTTP 请求的中间件。
// 本文件实现基于 JWT 的鉴权，提供三个中间件：
//   - Auth         ：必须携带合法令牌，否则 401(用于需要登录的接口)；
//   - OptionalAuth ：可选令牌，带了就解析身份，没带也放行(用于公开但想区分登录用户的接口)；
//   - RequireRole  ：在已登录的基础上进一步要求角色在白名单内(用于管理员接口)。
// 令牌校验通过后，会把 user_id 与 role 存入 gin.Context，供后续 handler 读取。
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

// 存入/读取 gin.Context 时使用的键名。集中定义可避免在多处手写字符串导致拼写不一致。
const (
	UserIDKey = "auth_user_id"
	RoleKey   = "auth_role"
)

// authClaims 是 JWT 载荷(payload)的结构：除自定义的 user_id、role 外，
// 还内嵌 jwt.RegisteredClaims 以复用标准字段(sub、签发时间、过期时间等)。
type authClaims struct {
	UserID uint64 `json:"user_id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}


// bearerToken 从 Authorization 请求头中提取令牌。
// 约定格式为 "Bearer <token>"，因此先校验前缀，再去掉前缀取出真正的令牌。
// 返回 error 是为了让调用方能区分“格式不对”，从而对强制登录接口返回 401。
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

// Auth 是“强制登录”中间件：令牌缺失、格式错误、服务端密钥未配置、验签失败或已过期时一律 401。
// 校验通过后把 user_id、role 写入 Context，再调用 c.Next() 放行给后续处理。
// 其中 WithValidMethods 限定只接受 HS256 算法，防止“算法混淆”类攻击；
// WithExpirationRequired 强制令牌必须带过期时间，避免出现永不过期的令牌。
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

// OptionalAuth 是“可选登录”中间件：令牌缺失或无效时直接放行(不报错)，
// 只有令牌有效时才解析并写入身份。用于“游客也能访问、但登录用户可看到更多内容”的接口。
func OptionalAuth(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil || jwtConfig.Secret == "" {
			c.Next()
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
			c.Next()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		c.Next()
	}
}

// CurrentRole 从 Context 读取当前用户角色。返回的 bool 表示是否取到(即是否已登录)。
// 注意从 Context 取出的值是 interface{}，需要做类型断言 role.(string) 还原成字符串。
func CurrentRole(c *gin.Context) (string, bool) {
	role, ok := c.Get(RoleKey)
	if !ok {
		return "", false
	}
	return role.(string), true
}

// RequireRole 生成一个“角色白名单”中间件：只有当前用户角色在 allowedRoles 中才放行。
// 通常与 Auth 配合使用(先用 Auth 确认已登录，再用本中间件校验角色)；
// 未登录或角色不符时返回 403(仅管理员可操作)。
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

// CurrentUserID 从 Context 读取当前用户 ID，bool 表示是否取到。
// 同样需要把 interface{} 断言回 uint64。
func CurrentUserID(c *gin.Context) (uint64, bool) { 
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	return userID.(uint64), true
}