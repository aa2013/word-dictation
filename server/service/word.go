package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/service/translator"
	"worddictation/server/util"
)

func ListWords(userId, libId int, q string, pageNum, pageSize int) (*dto.Page[dto.WordInfo], error) {
	if err := ensureMember(libId, userId); err != nil {
		return nil, err
	}
	pageNum, pageSize = normalizePage(pageNum, pageSize)
	q = strings.TrimSpace(q)

	query := global.DB.Table("word").
		Joins("JOIN lib_word lw ON lw.word_id = word.id AND lw.lib_id = ?", libId)
	if q != "" {
		query = query.Where("word.word LIKE ?", "%"+q+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []dto.WordInfo
	if err := query.Select("word.*").
		Order("word.id").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	attachDefaultExplains(userId, libId, rows)
	return dto.NewPage(rows, total, pageNum, pageSize), nil
}

func RandomWords(userId, libId int, size int) (*dto.Page[dto.WordInfo], error) {
	if err := ensureMember(libId, userId); err != nil {
		return nil, err
	}
	if size < 1 || size > 500 {
		return nil, util.BadRequest("数量必须在 1~500 之间")
	}
	var total int64
	global.DB.Table("word").
		Joins("JOIN lib_word lw ON lw.word_id = word.id AND lw.lib_id = ?", libId).
		Count(&total)
	if total < int64(size) {
		return nil, util.BadRequest("数量不足")
	}
	var rows []dto.WordInfo
	if err := global.DB.Table("word").Select("word.*").
		Joins("JOIN lib_word lw ON lw.word_id = word.id AND lw.lib_id = ?", libId).
		Order("RANDOM()").
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	attachDefaultExplains(userId, libId, rows)
	return dto.NewPage(rows, int64(len(rows)), 1, size), nil
}

// attachDefaultExplains 批量填充默认释义：用户自定义释义优先，否则取库内 is_default
func attachDefaultExplains(userId, libId int, rows []dto.WordInfo) {
	if len(rows) == 0 {
		return
	}
	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	var libExplains []model.WordExplain
	global.DB.Where("lib_id = ? AND is_default = 1 AND word_id IN ?", libId, ids).Find(&libExplains)
	byWord := make(map[int]model.WordExplain, len(libExplains))
	for _, e := range libExplains {
		byWord[e.WordId] = e
	}
	var customExplains []model.WordExplain
	global.DB.Table("word_explain_custom wec").
		Select("we.*").
		Joins("JOIN word_explain we ON we.id = wec.exp_id").
		Where("wec.user_id = ? AND wec.lib_id = ? AND wec.word_id IN ?", userId, libId, ids).
		Find(&customExplains)
	for _, e := range customExplains {
		byWord[e.WordId] = e
	}
	for i := range rows {
		if e, ok := byWord[rows[i].Id]; ok {
			rows[i].Explain = &dto.Explain{Explanation: e.Explanation, Type: e.Type}
		}
	}
}

func ImportWord(userId, libId int, wordStr string) (int, error) {
	wordStr = strings.TrimSpace(wordStr)
	if wordStr == "" {
		return 0, util.BadRequest("单词不能为空")
	}
	var lib model.Lib
	err := global.DB.First(&lib, libId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, util.NotFound("给定词库不存在，导入失败")
	}
	if err != nil {
		return 0, err
	}
	if lib.Creator != userId {
		return 0, util.Forbidden("只能导入到自己创建的词库中")
	}

	var existing model.Word
	err = global.DB.Where("word = ?", wordStr).First(&existing).Error
	wordExists := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	newWordId := 0
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		if !wordExists {
			tr, trErr := translator.Translate(wordStr)
			if trErr != nil {
				return util.BadGateway("单词查询失败：" + trErr.Error())
			}
			w := model.Word{
				Word:        wordStr,
				UsSymbol:    tr.UsSymbol,
				UsSymbolMp3: tr.UsMp3,
				EnSymbol:    tr.EnSymbol,
				EnSymbolMp3: tr.EnMp3,
			}
			if err := tx.Create(&w).Error; err != nil {
				return err
			}
			for i, e := range tr.Explains {
				base := model.WordExplain{
					WordId:      w.Id,
					Explanation: e.Explanation,
					Type:        e.Type,
					IsDefault:   i == 0,
				}
				toLib := base
				toLib.LibId = libId
				if err := tx.Create(&toLib).Error; err != nil {
					return err
				}
				toRoot := base
				toRoot.LibId = RootLibId
				if err := tx.Create(&toRoot).Error; err != nil {
					return err
				}
			}
			if err := tx.Create(&model.LibWord{LibId: RootLibId, WordId: w.Id}).Error; err != nil {
				return err
			}
			touchLib(tx, RootLibId)
			newWordId = w.Id
		} else {
			var src []model.WordExplain
			if err := tx.Where("lib_id = ? AND word_id = ?", RootLibId, existing.Id).Find(&src).Error; err != nil {
				return err
			}
			for _, e := range src {
				copy := e
				copy.Id = 0
				copy.LibId = libId
				if err := tx.Create(&copy).Error; err != nil {
					return err
				}
			}
			newWordId = existing.Id
		}
		if err := tx.Create(&model.LibWord{LibId: libId, WordId: newWordId}).Error; err != nil {
			if isUniqueViolation(err) {
				return util.Conflict("该单词已在此词库中")
			}
			return err
		}
		touchLib(tx, libId)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newWordId, nil
}

func GetExplains(userId, wordId, libId int) ([]dto.ExplainInfo, error) {
	if err := ensureMember(libId, userId); err != nil {
		return nil, err
	}
	var explains []model.WordExplain
	if err := global.DB.Where("lib_id = ? AND word_id = ?", libId, wordId).
		Order("id").Find(&explains).Error; err != nil {
		return nil, err
	}
	res := make([]dto.ExplainInfo, 0, len(explains))
	var custom model.WordExplainCustom
	err := global.DB.Where("user_id = ? AND lib_id = ? AND word_id = ?", userId, libId, wordId).
		First(&custom).Error
	customExpId := 0
	if err == nil {
		customExpId = custom.ExpId
	}
	for _, e := range explains {
		res = append(res, dto.ExplainInfo{
			Id:            e.Id,
			WordId:        e.WordId,
			LibId:         e.LibId,
			Explanation:   e.Explanation,
			Type:          e.Type,
			IsDefault:     e.IsDefault,
			CustomDefault: customExpId == e.Id,
		})
	}
	return res, nil
}

func ChangeDefaultExplain(userId, libId, wordId, explainId int) error {
	if err := ensureMember(libId, userId); err != nil {
		return err
	}
	var lib model.Lib
	if err := global.DB.First(&lib, libId).Error; err != nil {
		return util.NotFound("词库不存在")
	}
	var explain model.WordExplain
	err := global.DB.Where("id = ? AND word_id = ?", explainId, wordId).First(&explain).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("释义不存在")
	}
	if err != nil {
		return err
	}

	if lib.Creator == userId {
		// 创建者：修改库内公共默认释义
		return global.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.WordExplain{}).
				Where("word_id = ? AND lib_id = ?", wordId, libId).
				Update("is_default", false).Error; err != nil {
				return err
			}
			if err := tx.Model(&explain).Update("is_default", true).Error; err != nil {
				return err
			}
			touchLib(tx, libId)
			return nil
		})
	}
	// 普通成员：写入自定义默认释义
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var cnt int64
		tx.Model(&model.WordExplainCustom{}).
			Where("user_id = ? AND lib_id = ? AND word_id = ?", userId, libId, wordId).
			Count(&cnt)
		if cnt > 0 {
			return tx.Model(&model.WordExplainCustom{}).
				Where("user_id = ? AND lib_id = ? AND word_id = ?", userId, libId, wordId).
				Update("exp_id", explainId).Error
		}
		return tx.Create(&model.WordExplainCustom{
			UserId: userId, LibId: libId, WordId: wordId, ExpId: explainId,
		}).Error
	})
}

func normalizePage(pageNum, pageSize int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return pageNum, pageSize
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint")
}
