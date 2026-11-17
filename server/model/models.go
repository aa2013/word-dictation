package model

// User 用户表
type User struct {
	Id       int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Account  string `gorm:"column:account;not null" json:"account"`
	UserName string `gorm:"column:user_name;not null" json:"userName"`
	Password string `gorm:"column:password;not null" json:"-"`
}

func (User) TableName() string { return "user" }

// Session 登录会话（JWT jti），用于登出吊销
type Session struct {
	Id        string    `gorm:"primaryKey;column:id;size:64" json:"id"`
	UserId    int       `gorm:"column:user_id;index;not null" json:"userId"`
	CreatedAt LocalTime `gorm:"column:created_at;not null" json:"createdAt"`
	ExpiresAt LocalTime `gorm:"column:expires_at;index;not null" json:"expiresAt"`
}

func (Session) TableName() string { return "session" }

// Lib 词库表，id=1 为内置总词库
type Lib struct {
	Id         int        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	LibName    string     `gorm:"column:lib_name;not null" json:"libName"`
	Desc       string     `gorm:"column:desc;not null" json:"desc"`
	Cover      string     `gorm:"column:cover;not null" json:"cover"`
	Creator    int        `gorm:"column:creator;not null" json:"creator"`
	Common     bool       `gorm:"column:common;not null" json:"common"`
	CreateTime LocalTime  `gorm:"column:create_time;not null" json:"createTime"`
	UpdateTime *LocalTime `gorm:"column:update_time" json:"updateTime"`
}

func (Lib) TableName() string { return "lib" }

type LibWord struct {
	LibId  int `gorm:"primaryKey;autoIncrement:false;column:lib_id" json:"libId"`
	WordId int `gorm:"primaryKey;autoIncrement:false;column:word_id" json:"wordId"`
}

func (LibWord) TableName() string { return "lib_word" }

type UserLib struct {
	UserId int `gorm:"primaryKey;autoIncrement:false;column:user_id" json:"userId"`
	LibId  int `gorm:"primaryKey;autoIncrement:false;column:lib_id" json:"libId"`
}

func (UserLib) TableName() string { return "user_lib" }

type Word struct {
	Id          int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Word        string `gorm:"column:word;not null;unique" json:"word"`
	UsSymbol    string `gorm:"column:us_symbol" json:"usSymbol"`
	UsSymbolMp3 string `gorm:"column:us_symbol_mp3" json:"usSymbolMp3"`
	EnSymbol    string `gorm:"column:en_symbol" json:"enSymbol"`
	EnSymbolMp3 string `gorm:"column:en_symbol_mp3" json:"enSymbolMp3"`
}

func (Word) TableName() string { return "word" }

type WordExplain struct {
	Id          int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	WordId      int    `gorm:"column:word_id;not null" json:"wordId"`
	LibId       int    `gorm:"column:lib_id;not null" json:"libId"`
	Explanation string `gorm:"column:explanation;not null" json:"explanation"`
	Type        string `gorm:"column:type;not null" json:"type"`
	IsDefault   bool   `gorm:"column:is_default;not null" json:"isDefault"`
}

func (WordExplain) TableName() string { return "word_explain" }

// WordExplainCustom 用户在某词库中自定义的默认释义
type WordExplainCustom struct {
	UserId int `gorm:"primaryKey;autoIncrement:false;column:user_id" json:"userId"`
	LibId  int `gorm:"primaryKey;autoIncrement:false;column:lib_id" json:"libId"`
	WordId int `gorm:"primaryKey;autoIncrement:false;column:word_id" json:"wordId"`
	ExpId  int `gorm:"column:exp_id;not null" json:"expId"`
}

func (WordExplainCustom) TableName() string { return "word_explain_custom" }

type Plan struct {
	Id         int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserId     int       `gorm:"column:user_id;not null" json:"userId"`
	Name       string    `gorm:"column:name;not null" json:"name"`
	LibName    string    `gorm:"column:lib_name;not null" json:"libName"`
	Disorder   bool      `gorm:"column:disorder;not null" json:"disorder"`
	Ch2En      bool      `gorm:"column:ch2en;not null" json:"ch2en"`
	Repeat     int       `gorm:"column:repeat;not null" json:"repeat"`
	CreateTime LocalTime `gorm:"column:create_time;not null" json:"createTime"`
}

func (Plan) TableName() string { return "plan" }

type PlanWord struct {
	Id      int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlanId  int    `gorm:"column:plan_id;not null" json:"planId"`
	Word    string `gorm:"column:word;not null" json:"word"`
	Explain string `gorm:"column:explain;not null" json:"explain"`
}

func (PlanWord) TableName() string { return "plan_word" }
