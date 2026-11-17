package middleware

import (
	"errors"
	"log"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"

	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/util"
)

// Auth JWT + 会话吊销校验，通过后注入 uid / jti
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if token == "" {
			abort(c, util.Unauthorized("未登录"))
			return
		}
		claims, err := util.ParseJWT(global.Config.JWT.Secret, token)
		if err != nil {
			abort(c, util.Unauthorized("登录已过期，请重新登录"))
			return
		}
		var cnt int64
		global.DB.Model(&model.Session{}).
			Where("id = ? AND expires_at > ?", claims.JTI, model.Now()).
			Count(&cnt)
		if cnt == 0 {
			abort(c, util.Unauthorized("登录已失效，请重新登录"))
			return
		}
		c.Set("uid", claims.UID)
		c.Set("jti", claims.JTI)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("[panic] %v\n%s", err, debug.Stack())
		util.Fail(c, util.Internal("服务器内部错误"))
	})
}

func abort(c *gin.Context, err error) {
	var e *util.Err
	if !errors.As(err, &e) {
		e = util.Internal("服务器内部错误")
	}
	c.AbortWithStatusJSON(e.Status, util.Body{Code: e.Status, Message: e.Message, Data: nil})
}
