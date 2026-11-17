package handler

import (
	"github.com/gin-gonic/gin"

	"worddictation/server/dto"
	"worddictation/server/service"
	"worddictation/server/util"
)

func Login(c *gin.Context) {
	var req dto.LoginReq
	if !bindJSON(c, &req) {
		return
	}
	res, err := service.Login(req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func Logout(c *gin.Context) {
	if err := service.Logout(c.GetString("jti")); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, nil)
}

func Me(c *gin.Context) {
	res, err := service.Me(c.GetInt("uid"))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, res)
}

func Register(c *gin.Context) {
	var req dto.RegisterReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.Register(req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func SendRegisterCode(c *gin.Context) {
	var req dto.SendCodeReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.SendRegisterCodeMail(req.Email); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func SendResetCode(c *gin.Context) {
	var req dto.SendCodeReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.SendResetPwdCode(req.Email); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.ResetPassword(req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func UpdatePassword(c *gin.Context) {
	var req dto.UpdatePasswordReq
	if !bindJSON(c, &req) {
		return
	}
	if err := service.UpdatePassword(c.GetInt("uid"), req); err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, true)
}

func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		util.Fail(c, util.BadRequest("参数错误"))
		return false
	}
	return true
}
