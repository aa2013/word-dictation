package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"worddictation/server/dto"
	"worddictation/server/service"
	"worddictation/server/util"
)

// ListLibs scope=self 返回我的词库列表；scope=common 返回公开词库分页
func ListLibs(c *gin.Context) {
	uid := c.GetInt("uid")
	switch c.Query("scope") {
	case "self":
		res, err := service.GetSelfLibs(uid)
		if err != nil {
			util.Fail(c, err)
			return
		}
		util.OK(c, res)
	default:
		pageNum, pageSize := pageParams(c)
		res, err := service.GetCommonLibs(uid, c.Query("q"), pageNum, pageSize)
		if err != nil {
			util.Fail(c, err)
			return
		}
		util.OK(c, res)
	}
}

func GetLibInfo(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	res, err := service.GetLibInfo(c.GetInt("uid"), id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func CreateLib(c *gin.Context) {
	var req dto.SaveLibReq
	if !bindJSON(c, &req) {
		return
	}
	id, err := service.CreateLib(c.GetInt("uid"), req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"id": id})
}

func UpdateLib(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req dto.SaveLibReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.UpdateLib(c.GetInt("uid"), id, req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func DeleteLib(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteLib(c.GetInt("uid"), id); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func AdoptLib(c *gin.Context) {
	id, ok := pathInt(c, "id")
	if !ok {
		return
	}
	if err := service.AddLib(c.GetInt("uid"), id); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func pathInt(c *gin.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		util.Fail(c, util.BadRequest("参数错误"))
		return 0, false
	}
	return v, true
}

func pageParams(c *gin.Context) (int, int) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	return pageNum, pageSize
}
