package service

import (
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/service/storage"
	"worddictation/server/util"
)

const RootLibId = 1

func ensureMember(libId, userId int) error {
	var lib model.Lib
	err := global.DB.First(&lib, libId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("词库不存在")
	}
	if err != nil {
		return err
	}
	var cnt int64
	global.DB.Model(&model.UserLib{}).Where("user_id = ? AND lib_id = ?", userId, libId).Count(&cnt)
	if cnt == 0 {
		return util.Forbidden("你没有此库的访问权")
	}
	return nil
}

func CreateLib(userId int, req dto.SaveLibReq) (int, error) {
	libName := strings.TrimSpace(req.LibName)
	if libName == "" {
		return 0, util.BadRequest("词库名不能为空")
	}
	lib := model.Lib{
		LibName:    libName,
		Desc:       req.Desc,
		Cover:      "/covers/default.jpg",
		Creator:    userId,
		Common:     req.Common,
		CreateTime: model.Now(),
	}
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&lib).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserLib{UserId: userId, LibId: lib.Id}).Error
	})
	if err != nil {
		return 0, err
	}
	if util.IsBase64(req.Cover) {
		url, err := storage.UploadCover(lib.Id, req.Cover)
		if err != nil {
			log.Printf("[storage] 封面上传失败 lib=%d: %v", lib.Id, err)
		} else {
			global.DB.Model(&lib).Update("cover", url)
		}
	}
	return lib.Id, nil
}

func UpdateLib(userId, libId int, req dto.SaveLibReq) error {
	if libId == RootLibId {
		return util.Forbidden("不能修改根词库")
	}
	var lib model.Lib
	err := global.DB.First(&lib, libId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("词库不存在")
	}
	if err != nil {
		return err
	}
	if lib.Creator != userId {
		return util.Forbidden("无权限修改他人创建的词库")
	}
	libName := strings.TrimSpace(req.LibName)
	if libName == "" {
		return util.BadRequest("词库名不能为空")
	}
	updates := map[string]any{
		"lib_name":    libName,
		"desc":        req.Desc,
		"common":      req.Common,
		"update_time": model.Now(),
	}
	if util.IsBase64(req.Cover) {
		url, err := storage.UploadCover(libId, req.Cover)
		if err != nil {
			log.Printf("[storage] 封面上传失败 lib=%d: %v", libId, err)
		} else {
			updates["cover"] = url
		}
	}
	return global.DB.Model(&lib).Updates(updates).Error
}

func DeleteLib(userId, libId int) error {
	if libId == RootLibId {
		return util.Forbidden("不能删除根词库")
	}
	var cnt int64
	global.DB.Model(&model.UserLib{}).Where("user_id = ? AND lib_id = ?", userId, libId).Count(&cnt)
	if cnt == 0 {
		return util.NotFound("词库不存在或未添加")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND lib_id = ?", userId, libId).
			Delete(&model.UserLib{}).Error; err != nil {
			return err
		}
		var remain int64
		tx.Model(&model.UserLib{}).Where("lib_id = ?", libId).Count(&remain)
		if remain > 0 {
			return nil
		}
		if err := tx.Delete(&model.Lib{}, libId).Error; err != nil {
			return err
		}
		if err := tx.Where("lib_id = ?", libId).Delete(&model.LibWord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("lib_id = ?", libId).Delete(&model.WordExplain{}).Error; err != nil {
			return err
		}
		return tx.Where("lib_id = ?", libId).Delete(&model.WordExplainCustom{}).Error
	})
}

func AddLib(userId, libId int) error {
	var lib model.Lib
	err := global.DB.First(&lib, libId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("词库不存在")
	}
	if err != nil {
		return err
	}
	if !lib.Common {
		return util.Forbidden("非公共库，无权限")
	}
	var cnt int64
	global.DB.Model(&model.UserLib{}).Where("user_id = ? AND lib_id = ?", userId, libId).Count(&cnt)
	if cnt > 0 {
		return util.Conflict("已添加该词库")
	}
	return global.DB.Create(&model.UserLib{UserId: userId, LibId: libId}).Error
}

const libListQuery = `
SELECT lib.id, lib.lib_name, lib.desc, lib.cover, lib.creator, lib.common, lib.create_time, lib.update_time,
  u.user_name AS creator_name,
  (SELECT COUNT(*) FROM lib_word lw WHERE lw.lib_id = lib.id) AS cnt,
  (SELECT COUNT(*) FROM user_lib ul WHERE ul.lib_id = lib.id) AS uses
FROM lib
JOIN user u ON u.id = lib.creator
`

func GetSelfLibs(userId int) ([]dto.LibInfo, error) {
	libs := make([]dto.LibInfo, 0)
	sql := libListQuery + `
JOIN user_lib ul ON ul.lib_id = lib.id AND ul.user_id = ?
ORDER BY lib.id`
	if err := global.DB.Raw(sql, userId).Scan(&libs).Error; err != nil {
		return nil, err
	}
	for i := range libs {
		libs[i].Self = libs[i].Creator == userId
	}
	return libs, nil
}

func GetCommonLibs(userId int, q string, pageNum, pageSize int) (*dto.Page[dto.CommonLibInfo], error) {
	cond := "WHERE lib.common = 1"
	args := []any{}
	if strings.TrimSpace(q) != "" {
		cond += " AND lib.lib_name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(q)+"%")
	}
	var total int64
	if err := global.DB.Raw("SELECT COUNT(*) FROM lib "+cond, args...).Scan(&total).Error; err != nil {
		return nil, err
	}
	var libs []dto.LibInfo
	sql := libListQuery + cond + " ORDER BY lib.id LIMIT ? OFFSET ?"
	args = append(args, pageSize, (pageNum-1)*pageSize)
	if err := global.DB.Raw(sql, args...).Scan(&libs).Error; err != nil {
		return nil, err
	}
	var memberIds []int
	if err := global.DB.Model(&model.UserLib{}).Where("user_id = ?", userId).
		Pluck("lib_id", &memberIds).Error; err != nil {
		return nil, err
	}
	has := make(map[int]bool, len(memberIds))
	for _, id := range memberIds {
		has[id] = true
	}
	data := make([]dto.CommonLibInfo, 0, len(libs))
	for _, l := range libs {
		l.Self = l.Creator == userId
		data = append(data, dto.CommonLibInfo{LibInfo: l, HasLib: has[l.Id]})
	}
	return dto.NewPage(data, total, pageNum, pageSize), nil
}

func GetLibInfo(userId, libId int) (*dto.LibInfo, error) {
	if err := ensureMember(libId, userId); err != nil {
		return nil, err
	}
	var lib dto.LibInfo
	sql := libListQuery + "WHERE lib.id = ?"
	if err := global.DB.Raw(sql, libId).Scan(&lib).Error; err != nil {
		return nil, err
	}
	lib.Self = lib.Creator == userId
	return &lib, nil
}

func touchLib(tx *gorm.DB, libId int) {
	tx.Model(&model.Lib{}).Where("id = ?", libId).Update("update_time", model.Now())
}
