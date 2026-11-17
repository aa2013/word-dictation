package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"worddictation/server/global"
	"worddictation/server/handler"
	"worddictation/server/middleware"
	"worddictation/server/util"
	"worddictation/server/web"
)

func Setup() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery(), middleware.CORS())

	api := r.Group("/api")
	{
		api.GET("/setup/status", handler.SetupStatus)
		api.POST("/setup/init", handler.InitSystem)
		api.POST("/auth/login", handler.Login)
		api.POST("/users", handler.Register)
		api.POST("/users/codes/register", handler.SendRegisterCode)
		api.POST("/users/codes/reset", handler.SendResetCode)
		api.POST("/users/password/reset", handler.ResetPassword)
	}
	auth := api.Group("", middleware.Auth())
	{
		auth.POST("/auth/logout", handler.Logout)
		auth.GET("/auth/me", handler.Me)
		auth.PUT("/users/password", handler.UpdatePassword)

		auth.GET("/libs", handler.ListLibs)
		auth.POST("/libs", handler.CreateLib)
		auth.GET("/libs/:id", handler.GetLibInfo)
		auth.PUT("/libs/:id", handler.UpdateLib)
		auth.DELETE("/libs/:id", handler.DeleteLib)
		auth.POST("/libs/:id/adopt", handler.AdoptLib)
		auth.GET("/libs/:id/words", handler.ListWords)
		auth.POST("/libs/:id/words", handler.ImportWord)
		auth.GET("/libs/:id/random-words", handler.RandomWords)
		auth.PUT("/libs/:id/default-explain", handler.ChangeDefaultExplain)
		auth.GET("/words/:wordId/explains", handler.GetExplains)

		auth.GET("/plans", handler.ListPlans)
		auth.POST("/plans", handler.CreatePlan)
		auth.GET("/plans/:id/words", handler.GetPlanWords)
		auth.DELETE("/plans/:id", handler.DeletePlan)
	}

	r.Static("/covers", global.Config.Storage.Dir)
	registerStatic(r)
	return r
}

// registerStatic 服务嵌入的前端 SPA：命中文件直接返回，否则回退 index.html（history 路由）
func registerStatic(r *gin.Engine) {
	distFS, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(distFS))
	index, indexErr := fs.ReadFile(distFS, "index.html")

	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/covers/") {
			util.Fail(c, util.NotFound("资源不存在"))
			return
		}
		if indexErr == nil && (p == "/" || p == "") {
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
			return
		}
		if f, err := distFS.Open(strings.TrimPrefix(p, "/")); err == nil {
			if stat, err := f.Stat(); err == nil && !stat.IsDir() {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
			f.Close()
		}
		if indexErr == nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
			return
		}
		c.String(http.StatusNotFound, "not found")
	})
}
