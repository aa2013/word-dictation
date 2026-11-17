package service

import (
	"errors"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/util"
)

const sysInitializedKey = "sys_initialized"

var setupMu sync.Mutex

// IsInitialized 无标记时视为已初始化，避免缺记录的库开放初始化接口
func IsInitialized() (bool, error) {
	var s model.Setting
	err := global.DB.Where("`key` = ?", sysInitializedKey).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return s.Value == "1", nil
}

// InitSystem 首次部署初始化：更新内置管理员账号密码并落初始化标记。
// 已初始化的库一律拒绝，防止接口被外部扫描利用。
func InitSystem(req dto.SetupReq) error {
	setupMu.Lock()
	defer setupMu.Unlock()

	initialized, err := IsInitialized()
	if err != nil {
		return err
	}
	if initialized {
		return util.Forbidden("系统已初始化")
	}
	account := strings.TrimSpace(req.Account)
	if account == "" {
		return util.BadRequest("管理员账号不能为空")
	}
	if req.Password == "" {
		return util.BadRequest("密码不能为空")
	}
	if req.Password != req.PasswordAgain {
		return util.BadRequest("两次密码不一致")
	}
	var cnt int64
	global.DB.Model(&model.User{}).Where("account = ? AND id <> 1", account).Count(&cnt)
	if cnt > 0 {
		return util.Conflict("该账号已被使用")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.User{}).Where("id = ?", 1).
			Updates(map[string]any{"account": account, "password": string(hash)})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return util.NotFound("管理员账号不存在")
		}
		return tx.Model(&model.Setting{}).Where("`key` = ?", sysInitializedKey).
			Update("value", "1").Error
	})
}
