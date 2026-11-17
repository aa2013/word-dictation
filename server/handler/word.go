package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"worddictation/server/dto"
	"worddictation/server/service"
	"worddictation/server/util"
)

func ListWords(c *gin.Context) {
	libId, ok := pathInt(c, "id")
	if !ok {
		return
	}
	pageNum, pageSize := pageParams(c)
	res, err := service.ListWords(c.GetInt("uid"), libId, c.Query("q"), pageNum, pageSize)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func RandomWords(c *gin.Context) {
	libId, ok := pathInt(c, "id")
	if !ok {
		return
	}
	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		util.Fail(c, util.BadRequest("参数错误"))
		return
	}
	res, err := service.RandomWords(c.GetInt("uid"), libId, size)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func ImportWord(c *gin.Context) {
	libId, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req dto.ImportWordReq
	if !bindJSON(c, &req) {
		return
	}
	wordId, err := service.ImportWord(c.GetInt("uid"), libId, req.Word)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"wordId": wordId})
}

func GetExplains(c *gin.Context) {
	wordId, ok := pathInt(c, "wordId")
	if !ok {
		return
	}
	libId, err := strconv.Atoi(c.Query("libId"))
	if err != nil {
		util.Fail(c, util.BadRequest("参数错误"))
		return
	}
	res, err := service.GetExplains(c.GetInt("uid"), wordId, libId)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func ChangeDefaultExplain(c *gin.Context) {
	libId, ok := pathInt(c, "id")
	if !ok {
		return
	}
	var req dto.ChangeDefaultExplainReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.ChangeDefaultExplain(c.GetInt("uid"), libId, req.WordId, req.ExplainId); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}
