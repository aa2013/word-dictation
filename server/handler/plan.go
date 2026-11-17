package handler

import (
	"github.com/gin-gonic/gin"

	"worddictation/server/dto"
	"worddictation/server/service"
	"worddictation/server/util"
)

func ListPlans(c *gin.Context) {
	pageNum, pageSize := pageParams(c)
	res, err := service.ListPlans(c.GetInt("uid"), c.Query("q"), pageNum, pageSize)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func CreatePlan(c *gin.Context) {
	var req dto.CreatePlanReq
	if !bindJSON(c, &req) {
		return
	}
	id, err := service.CreatePlan(c.GetInt("uid"), req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"id": id})
}

func GetPlanWords(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	res, err := service.GetPlanWords(c.GetInt("uid"), id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func DeletePlan(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	if err := service.DeletePlan(c.GetInt("uid"), id); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}
