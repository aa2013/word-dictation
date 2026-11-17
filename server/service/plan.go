package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"worddictation/server/dto"
	"worddictation/server/global"
	"worddictation/server/model"
	"worddictation/server/util"
)

func ListPlans(userId int, q string, pageNum, pageSize int) (*dto.Page[dto.PlanInfo], error) {
	pageNum, pageSize = normalizePage(pageNum, pageSize)
	query := global.DB.Model(&model.Plan{}).Where("user_id = ?", userId)
	q = strings.TrimSpace(q)
	if q != "" {
		query = query.Where("name LIKE ? OR lib_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var plans []model.Plan
	if err := query.Order("id").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&plans).Error; err != nil {
		return nil, err
	}
	cntByPlan := planWordCnt(plans)
	data := make([]dto.PlanInfo, 0, len(plans))
	for _, p := range plans {
		data = append(data, dto.PlanInfo{
			Id:         p.Id,
			UserId:     p.UserId,
			Name:       p.Name,
			LibName:    p.LibName,
			Disorder:   p.Disorder,
			Ch2En:      p.Ch2En,
			Repeat:     p.Repeat,
			CreateTime: p.CreateTime,
			Cnt:        cntByPlan[p.Id],
		})
	}
	return dto.NewPage(data, total, pageNum, pageSize), nil
}

func planWordCnt(plans []model.Plan) map[int]int {
	res := make(map[int]int, len(plans))
	if len(plans) == 0 {
		return res
	}
	ids := make([]int, 0, len(plans))
	for _, p := range plans {
		ids = append(ids, p.Id)
	}
	type row struct {
		PlanId int
		Cnt    int
	}
	var rows []row
	global.DB.Model(&model.PlanWord{}).
		Select("plan_id, COUNT(*) AS cnt").
		Where("plan_id IN ?", ids).
		Group("plan_id").
		Scan(&rows)
	for _, r := range rows {
		res[r.PlanId] = r.Cnt
	}
	return res
}

func CreatePlan(userId int, req dto.CreatePlanReq) (int, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return 0, util.BadRequest("方案名不能为空")
	}
	if len(req.Words) == 0 {
		return 0, util.BadRequest("单词不能为空")
	}
	if req.Repeat < 1 {
		return 0, util.BadRequest("重复次数不能小于 1")
	}
	var lib model.Lib
	err := global.DB.First(&lib, req.LibId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, util.NotFound("词库不存在")
	}
	if err != nil {
		return 0, err
	}
	plan := model.Plan{
		UserId:     userId,
		Name:       req.Name,
		LibName:    lib.LibName,
		Disorder:   req.Disorder,
		Ch2En:      req.Ch2En,
		Repeat:     req.Repeat,
		CreateTime: model.Now(),
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&plan).Error; err != nil {
			return err
		}
		for _, w := range req.Words {
			if strings.TrimSpace(w.Word) == "" {
				continue
			}
			if err := tx.Create(&model.PlanWord{
				PlanId:  plan.Id,
				Word:    w.Word,
				Explain: w.Explain,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return plan.Id, nil
}

func GetPlanWords(userId, planId int) ([]dto.PlanWordItem, error) {
	var plan model.Plan
	err := global.DB.First(&plan, planId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, util.NotFound("方案不存在")
	}
	if err != nil {
		return nil, err
	}
	if plan.UserId != userId {
		return nil, util.Forbidden("无权限查看该方案")
	}
	var words []model.PlanWord
	if err := global.DB.Where("plan_id = ?", planId).Order("id").Find(&words).Error; err != nil {
		return nil, err
	}
	res := make([]dto.PlanWordItem, 0, len(words))
	for _, w := range words {
		res = append(res, dto.PlanWordItem{Word: w.Word, Explain: w.Explain})
	}
	return res, nil
}

func DeletePlan(userId, planId int) error {
	var plan model.Plan
	err := global.DB.First(&plan, planId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.NotFound("方案不存在")
	}
	if err != nil {
		return err
	}
	if plan.UserId != userId {
		return util.Forbidden("无权限删除他人创建的内容")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&plan).Error; err != nil {
			return err
		}
		return tx.Where("plan_id = ?", planId).Delete(&model.PlanWord{}).Error
	})
}
