package middleware

import (
	"tomatoBlogDB/model"

	"github.com/kataras/iris/v12"
)

func RequireAdmin(c iris.Context) {
	role, _ := c.Values().GetInt("current_user_role")
	if role != 999 {
		model.Fail(c, model.ResponseJson{Code: 403, Msg: "Insufficient permissions"})
		c.StopExecution()
		return
	}
	c.Next()
}
