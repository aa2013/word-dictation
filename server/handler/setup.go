package handler

import (
	"github.com/gin-gonic/gin"

	"worddictation/server/dto"
	"worddictation/server/service"
	"worddictation/server/util"
)

func SetupStatus(c *gin.Context) {
	initialized, err := service.IsInitialized()
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"initialized": initialized})
}

func InitSystem(c *gin.Context) {
	var req dto.SetupReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.InitSystem(req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}
