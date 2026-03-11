package middleware

import (
	"net/http"
	"strings"
	"tomatoBlogDB/model"
	"tomatoBlogDB/utils"

	"github.com/kataras/iris/v12"
)

const (
	ERR_CODE_INVALID_TOKEN = 10401 // Invalid token error code
	TOKEN_NAME             = "Authorization"
	TOKEN_PREFIX_KEY       = "Bearer" // Token prefix, can be customized
)

func tokenErr(c iris.Context, msg string) {
	if msg == "" {
		msg = "Invalid Token or Token Expired"
	}
	model.Fail(c, model.ResponseJson{
		Code:   ERR_CODE_INVALID_TOKEN,
		Msg:    msg,
		Status: http.StatusUnauthorized,
	})
	// = stop successive handler
	c.StopExecution()
}

// Unauthorized -- 401
// Dual track Authorization
func Auth() func(c iris.Context) {
	return func(c iris.Context) {
		var tokenString string
		// 1.1 try to phrase Header
		authHeader := c.GetHeader(TOKEN_NAME)

		// 1.2 phrase "Bearer <token>"
		// - strings.Fields will tackle with multiple blank spaces
		if authHeader != "" {
			parts := strings.Fields(authHeader)
			if parts[0] == "Bearer" {
				if len(parts) != 2 || parts[0] != TOKEN_PREFIX_KEY {
					tokenErr(c, "Invalid Token Format (expected `Bearer <token>`) but receive "+parts[0])
					return
				}
				tokenString = parts[1]
			}
		}

		// 1.3 phrase Cookie
		if tokenString == "" {
			tokenString = c.GetCookie("access_token")
		}

		if tokenString == "" {
			tokenErr(c, "Missing Authorization Header")
			return
		}

		// 2. phrase token, utils.ParseToken() will check whether token is expired.
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			tokenErr(c, "Token Invalid: "+err.Error())
			return
		}

		// 3. other checking
		if claims.ID <= 0 {
			tokenErr(c, "Token contains invalid User ID")
			return
		}

		// 4. inject UserID into context
		// - Subsequent will get info from c.Value().GetUint64("current_user_id")
		c.Values().Set("current_user_id", claims.ID)
		c.Values().Set("current_user_name", claims.Name)
		c.Values().Set("current_user_role", claims.Role)
		// 5. pass through
		c.Next()

	}
}
