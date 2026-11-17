package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/util"
)

func Register(req dto.RegisterReq) error {
	email := strings.TrimSpace(req.Email)
	if !util.IsEmail(email) {
		return util.BadRequest("邮箱格式不正确")
	}
	var cnt int64
	global.DB.Model(&model.User{}).Where("account = ?", email).Count(&cnt)
	if cnt > 0 {
		return util.Conflict("该邮箱已注册，请直接登录")
	}
	if req.Password == "" {
		return util.BadRequest("密码不能为空")
	}
	if req.Password != req.PasswordAgain {
		return util.BadRequest("两次密码不一致")
	}
	if !VerifyRegisterCode(email, strings.TrimSpace(req.Code)) {
		return util.BadRequest("验证码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var total int64
	global.DB.Model(&model.User{}).Count(&total)
	user := model.User{
		Account:  email,
		UserName: fmt.Sprintf("注册用户%d", total),
		Password: string(hash),
	}
	return global.DB.Create(&user).Error
}

func SendResetPwdCode(email string) error {
	var cnt int64
	global.DB.Model(&model.User{}).Where("account = ?", email).Count(&cnt)
	if cnt == 0 {
		return util.NotFound("用户不存在，请先注册")
	}
	return SendResetPwdCodeMail(email)
}

func ResetPassword(req dto.ResetPasswordReq) error {
	email := strings.TrimSpace(req.Email)
	var user model.User
	err := global.DB.Where("account = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("用户不存在，请先注册")
	}
	if err != nil {
		return err
	}
	if !VerifyResetPwdCode(email, strings.TrimSpace(req.Code)) {
		return util.BadRequest("验证码错误")
	}
	if req.Password == "" {
		return util.BadRequest("密码不能为空")
	}
	if req.Password != req.PasswordAgain {
		return util.BadRequest("两次密码不一致")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return global.DB.Save(&user).Error
}

func UpdatePassword(uid int, req dto.UpdatePasswordReq) error {
	var user model.User
	if err := global.DB.First(&user, uid).Error; err != nil {
		return err
	}
	if req.NewPassword != req.NewPasswordAgain {
		return util.BadRequest("两次密码不一致")
	}
	if req.NewPassword == "" {
		return util.BadRequest("密码不能为空")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)) != nil {
		return util.BadRequest("原密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return global.DB.Save(&user).Error
}
