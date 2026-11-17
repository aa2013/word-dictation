package service

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/util"
)

func Login(req dto.LoginReq) (*dto.LoginResult, error) {
	account := strings.TrimSpace(req.Account)
	if account == "" || req.Password == "" {
		return nil, util.BadRequest("账号或密码不能为空")
	}
	var user model.User
	err := global.DB.Where("account = ?", account).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.Unauthorized("账号或密码错误")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, util.Unauthorized("账号或密码错误")
	}

	days := global.Config.JWT.TokenDays
	if req.Remember {
		days = global.Config.JWT.RememberDays
	}
	if days <= 0 {
		days = 7
	}
	ttl := time.Duration(days) * 24 * time.Hour

	jti := util.RandomHex(16)
	token, err := util.SignJWT(global.Config.JWT.Secret, util.Claims{
		UID:      user.Id,
		Account:  user.Account,
		UserName: user.UserName,
		JTI:      jti,
	}, ttl)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 顺带清理过期会话
		if err := tx.Where("expires_at < ?", model.NewLocalTime(now)).Delete(&model.Session{}).Error; err != nil {
			return err
		}
		return tx.Create(&model.Session{
			Id:        jti,
			UserId:    user.Id,
			CreatedAt: model.NewLocalTime(now),
			ExpiresAt: model.NewLocalTime(now.Add(ttl)),
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return &dto.LoginResult{
		Token:     token,
		ExpiresIn: int64(ttl.Seconds()),
		User:      dto.UserInfo{Id: user.Id, Account: user.Account, UserName: user.UserName},
	}, nil
}

func Logout(jti string) error {
	if jti == "" {
		return nil
	}
	return global.DB.Where("id = ?", jti).Delete(&model.Session{}).Error
}

func Me(uid int) (*dto.UserInfo, error) {
	var user model.User
	if err := global.DB.First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.Unauthorized("用户不存在")
		}
		return nil, err
	}
	return &dto.UserInfo{Id: user.Id, Account: user.Account, UserName: user.UserName}, nil
}
