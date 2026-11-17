package dto

import "worddictation/server/model"

type LoginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type RegisterReq struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
	Code          string `json:"code"`
}

type SendCodeReq struct {
	Email string `json:"email"`
}

type ResetPasswordReq struct {
	Email         string `json:"email"`
	Code          string `json:"code"`
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
}

type UpdatePasswordReq struct {
	OldPassword      string `json:"oldPassword"`
	NewPassword      string `json:"newPassword"`
	NewPasswordAgain string `json:"newPasswordAgain"`
}

type SetupReq struct {
	Account       string `json:"account"`
	Password      string `json:"password"`
	PasswordAgain string `json:"passwordAgain"`
}

type SaveLibReq struct {
	LibName string `json:"libName"`
	Desc    string `json:"desc"`
	Cover   string `json:"cover"`
	Common  bool   `json:"common"`
}

type ImportWordReq struct {
	Word string `json:"word"`
}

type ChangeDefaultExplainReq struct {
	WordId    int `json:"wordId"`
	ExplainId int `json:"explainId"`
}

type PlanWordItem struct {
	Word    string `json:"word"`
	Explain string `json:"explain"`
}

type CreatePlanReq struct {
	Name     string         `json:"name"`
	LibId    int            `json:"libId"`
	Disorder bool           `json:"disorder"`
	Ch2En    bool           `json:"ch2en"`
	Repeat   int            `json:"repeat"`
	Words    []PlanWordItem `json:"words"`
}

type UserInfo struct {
	Id       int    `json:"id"`
	Account  string `json:"account"`
	UserName string `json:"userName"`
}

type LoginResult struct {
	Token     string   `json:"token"`
	ExpiresIn int64    `json:"expiresIn"`
	User      UserInfo `json:"user"`
}

type Explain struct {
	Explanation string `json:"explanation"`
	Type        string `json:"type"`
}

type WordInfo struct {
	Id          int      `json:"id"`
	Word        string   `json:"word"`
	UsSymbol    string   `json:"usSymbol"`
	UsSymbolMp3 string   `json:"usSymbolMp3"`
	EnSymbol    string   `json:"enSymbol"`
	EnSymbolMp3 string   `json:"enSymbolMp3"`
	Explain     *Explain `json:"explain" gorm:"-"`
}

type ExplainInfo struct {
	Id            int    `json:"id"`
	WordId        int    `json:"wordId"`
	LibId         int    `json:"libId"`
	Explanation   string `json:"explanation"`
	Type          string `json:"type"`
	IsDefault     bool   `json:"isDefault"`
	CustomDefault bool   `json:"customDefault"`
}

type LibInfo struct {
	Id          int              `json:"id"`
	LibName     string           `json:"libName"`
	Desc        string           `json:"desc"`
	Cover       string           `json:"cover"`
	CreatorName string           `json:"creatorName"`
	Creator     int              `json:"creator"`
	CreateTime  *model.LocalTime `json:"createTime"`
	UpdateTime  *model.LocalTime `json:"updateTime"`
	Common      bool             `json:"common"`
	Uses        int              `json:"uses"`
	Cnt         int              `json:"cnt"`
	Self        bool             `json:"self"`
}

type CommonLibInfo struct {
	LibInfo
	HasLib bool `json:"hasLib"`
}

type PlanInfo struct {
	Id         int             `json:"id"`
	UserId     int             `json:"userId"`
	Name       string          `json:"name"`
	LibName    string          `json:"libName"`
	Disorder   bool            `json:"disorder"`
	Ch2En      bool            `json:"ch2en"`
	Repeat     int             `json:"repeat"`
	CreateTime model.LocalTime `json:"createTime"`
	Cnt        int             `json:"cnt"`
}

type Page[T any] struct {
	List        []T   `json:"list"`
	Total       int64 `json:"total"`
	PageNum     int   `json:"pageNum"`
	PageSize    int   `json:"pageSize"`
	Pages       int64 `json:"pages"`
	HasNextPage bool  `json:"hasNextPage"`
	HasPrevPage bool  `json:"hasPreviousPage"`
}

func NewPage[T any](list []T, total int64, pageNum, pageSize int) *Page[T] {
	if list == nil {
		list = []T{}
	}
	pages := (total + int64(pageSize) - 1) / int64(pageSize)
	return &Page[T]{
		List:        list,
		Total:       total,
		PageNum:     pageNum,
		PageSize:    pageSize,
		Pages:       pages,
		HasNextPage: int64(pageNum) < pages,
		HasPrevPage: pageNum > 1,
	}
}
