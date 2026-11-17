package router_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"worddictation/server/config"
	"worddictation/server/global"
	"worddictation/server/migration"
	"worddictation/server/model"
	"worddictation/server/router"
	"worddictation/server/service"
	"worddictation/server/service/translator"
	"worddictation/server/util"
)

type fakeTranslator struct{}

func (fakeTranslator) Name() string { return "fake" }

func (fakeTranslator) Translate(word string) (translator.Result, error) {
	if word == "failme" {
		return translator.Result{}, errors.New("查询源不可用")
	}
	return translator.Result{
		EnSymbol: "/ˈæpl/",
		UsSymbol: "/ˈæpəl/",
		Explains: []translator.Explain{
			{Explanation: "苹果", Type: "n."},
			{Explanation: "苹果公司", Type: "n."},
		},
	}, nil
}

type testEnv struct {
	ts    *httptest.Server
	db    *gorm.DB
	token string
}

func setupEnv(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "t.db")), &gorm.Config{
		Logger: util.GormLogger(),
	})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	global.DB = db
	if err := migration.Run(db); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	global.Config = &config.Config{
		Port:    0,
		DB:      config.DB{Path: filepath.Join(dir, "t.db")},
		Storage: config.Storage{Dir: filepath.Join(dir, "storage")},
		JWT:     config.JWT{Secret: "test-secret", TokenDays: 1, RememberDays: 7},
		Email:   config.Email{Debug: true},
	}
	translator.SetTranslators([]translator.Translator{fakeTranslator{}})

	hash, _ := bcrypt.GenerateFromPassword([]byte("Passw0rd!"), bcrypt.MinCost)
	if err := db.Create(&model.User{Account: "tester@example.com", UserName: "测试用户", Password: string(hash)}).Error; err != nil {
		t.Fatal(err)
	}
	hash2, _ := bcrypt.GenerateFromPassword([]byte("Passw0rd!"), bcrypt.MinCost)
	if err := db.Create(&model.User{Account: "other@example.com", UserName: "其他用户", Password: string(hash2)}).Error; err != nil {
		t.Fatal(err)
	}

	r := router.Setup()
	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)

	env := &testEnv{ts: ts, db: db}
	res := env.do(t, http.MethodPost, "/api/auth/login",
		map[string]any{"account": "tester@example.com", "password": "Passw0rd!"}, "")
	if res.status != 200 {
		t.Fatalf("登录失败: %s", res.body)
	}
	env.token = res.data["token"].(string)
	return env
}

func (e *testEnv) do(t *testing.T, method, path string, body any, token string) testResponse {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, e.ts.URL+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求 %s %s 失败: %v", method, path, err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	out := testResponse{status: resp.StatusCode, body: buf.String()}
	_ = json.Unmarshal(buf.Bytes(), &out.envelope)
	if m, ok := out.envelope["data"].(map[string]any); ok {
		out.data = m
	}
	return out
}

type testResponse struct {
	status   int
	body     string
	envelope map[string]any
	data     map[string]any
}

func (r testResponse) num(t *testing.T, key string) int {
	t.Helper()
	v, ok := r.data[key].(float64)
	if !ok {
		t.Fatalf("响应 %d 缺少数值字段 %q: %s", r.status, key, r.body)
	}
	return int(v)
}

func TestAuthFlow(t *testing.T) {
	env := setupEnv(t)

	// 密码错误 → 401
	res := env.do(t, http.MethodPost, "/api/auth/login",
		map[string]any{"account": "tester@example.com", "password": "wrong"}, "")
	if res.status != 401 {
		t.Errorf("错误密码状态码 = %d, 期望 401", res.status)
	}

	// 未带 token → 401
	res = env.do(t, http.MethodGet, "/api/libs?scope=self", nil, "")
	if res.status != 401 {
		t.Errorf("未登录状态码 = %d, 期望 401", res.status)
	}

	// me
	res = env.do(t, http.MethodGet, "/api/auth/me", nil, env.token)
	if res.status != 200 || res.data["account"] != "tester@example.com" {
		t.Errorf("/auth/me 异常: %s", res.body)
	}

	// 登出后 token 失效
	res = env.do(t, http.MethodPost, "/api/auth/logout", nil, env.token)
	if res.status != 200 {
		t.Errorf("登出失败: %s", res.body)
	}
	res = env.do(t, http.MethodGet, "/api/auth/me", nil, env.token)
	if res.status != 401 {
		t.Errorf("登出后仍可访问: %d", res.status)
	}
}

func TestRegisterFlow(t *testing.T) {
	env := setupEnv(t)

	env.do(t, http.MethodPost, "/api/users/codes/register", map[string]any{"email": "new@example.com"}, "")
	code := service.GetCode("register", "new@example.com")
	if code == "" {
		t.Fatal("验证码未生成")
	}

	// 验证码错误 → 400
	res := env.do(t, http.MethodPost, "/api/users", map[string]any{
		"email": "new@example.com", "password": "Passw0rd!", "passwordAgain": "Passw0rd!", "code": "000000"}, "")
	if res.status != 400 {
		t.Errorf("错误验证码状态码 = %d, 期望 400", res.status)
	}

	res = env.do(t, http.MethodPost, "/api/users", map[string]any{
		"email": "new@example.com", "password": "Passw0rd!", "passwordAgain": "Passw0rd!", "code": code}, "")
	if res.status != 200 {
		t.Errorf("注册失败: %s", res.body)
	}

	// 重复注册 → 409
	res = env.do(t, http.MethodPost, "/api/users", map[string]any{
		"email": "new@example.com", "password": "Passw0rd!", "passwordAgain": "Passw0rd!", "code": code}, "")
	if res.status != 409 {
		t.Errorf("重复注册状态码 = %d, 期望 409", res.status)
	}

	// 新用户登录
	res = env.do(t, http.MethodPost, "/api/auth/login",
		map[string]any{"account": "new@example.com", "password": "Passw0rd!", "remember": true}, "")
	if res.status != 200 {
		t.Errorf("新用户登录失败: %s", res.body)
	}
}

func TestLibAndWordFlow(t *testing.T) {
	env := setupEnv(t)

	// 创建词库
	res := env.do(t, http.MethodPost, "/api/libs", map[string]any{
		"libName": "我的测试库", "desc": "测试", "common": true}, env.token)
	if res.status != 200 {
		t.Fatalf("创建词库失败: %s", res.body)
	}
	libId := res.num(t, "id")

	// 我的词库列表应包含且 cnt=0
	res = env.do(t, http.MethodGet, "/api/libs?scope=self", nil, env.token)
	if res.status != 200 || !strings.Contains(res.body, "我的测试库") {
		t.Fatalf("我的词库列表异常: %s", res.body)
	}
	// 公开词库应包含新库且 hasLib=true
	res = env.do(t, http.MethodGet, "/api/libs?scope=common", nil, env.token)
	if res.status != 200 || !strings.Contains(res.body, `"hasLib":true`) {
		t.Errorf("公开词库列表异常: %s", res.body)
	}

	// 导入单词（fake 翻译器，种子里不存在的词）
	res = env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "testword"}, env.token)
	if res.status != 200 {
		t.Fatalf("导入单词失败: %s", res.body)
	}
	wordId := res.num(t, "wordId")

	// 总词库应有释义副本
	var cnt int64
	env.db.Model(&model.WordExplain{}).Where("lib_id = ? AND word_id = ?", 1, wordId).Count(&cnt)
	if cnt != 2 {
		t.Errorf("总词库释义数 = %d, 期望 2", cnt)
	}

	// 重复导入 → 409
	res = env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "testword"}, env.token)
	if res.status != 409 {
		t.Errorf("重复导入状态码 = %d, 期望 409", res.status)
	}

	// 翻译失败 → 502
	res = env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "failme"}, env.token)
	if res.status != 502 {
		t.Errorf("翻译失败状态码 = %d, 期望 502", res.status)
	}

	// 单词列表带默认释义
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/libs/%d/words", libId), nil, env.token)
	if res.status != 200 || !strings.Contains(res.body, "苹果") {
		t.Errorf("单词列表异常: %s", res.body)
	}

	// 词库详情
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/libs/%d", libId), nil, env.token)
	if res.status != 200 || res.data["cnt"].(float64) != 1 {
		t.Errorf("词库详情异常: %s", res.body)
	}

	// 其他用户无权访问
	otherRes := env.do(t, http.MethodPost, "/api/auth/login",
		map[string]any{"account": "other@example.com", "password": "Passw0rd!"}, "")
	otherToken := otherRes.data["token"].(string)
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/libs/%d", libId), nil, otherToken)
	if res.status != 403 {
		t.Errorf("非成员访问状态码 = %d, 期望 403", res.status)
	}

	// 非创建者导入 → 403
	res = env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "pear"}, otherToken)
	if res.status != 403 {
		t.Errorf("非创建者导入状态码 = %d, 期望 403", res.status)
	}

	// 删除词库后 404
	res = env.do(t, http.MethodDelete, fmt.Sprintf("/api/libs/%d", libId), nil, env.token)
	if res.status != 200 {
		t.Fatalf("删除词库失败: %s", res.body)
	}
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/libs/%d", libId), nil, env.token)
	if res.status != 404 {
		t.Errorf("删除后访问状态码 = %d, 期望 404", res.status)
	}
}

func TestChangeDefaultExplain(t *testing.T) {
	env := setupEnv(t)

	res := env.do(t, http.MethodPost, "/api/libs", map[string]any{"libName": "释义库"}, env.token)
	libId := res.num(t, "id")
	res = env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "testword"}, env.token)
	wordId := res.num(t, "wordId")

	// 取该库全部释义
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/words/%d/explains?libId=%d", wordId, libId), nil, env.token)
	var explains []map[string]any
	data, _ := json.Marshal(res.envelope["data"])
	json.Unmarshal(data, &explains)
	if len(explains) != 2 {
		t.Fatalf("释义数 = %d, 期望 2", len(explains))
	}
	second := int(explains[1]["id"].(float64))

	// 切换默认释义（创建者路径）
	res = env.do(t, http.MethodPut, fmt.Sprintf("/api/libs/%d/default-explain", libId),
		map[string]any{"wordId": wordId, "explainId": second}, env.token)
	if res.status != 200 {
		t.Fatalf("切换默认释义失败: %s", res.body)
	}
	var def model.WordExplain
	env.db.Where("word_id = ? AND lib_id = ? AND is_default = 1", wordId, libId).First(&def)
	if def.Id != second {
		t.Errorf("默认释义 = %d, 期望 %d", def.Id, second)
	}
}

func TestPlanFlow(t *testing.T) {
	env := setupEnv(t)

	res := env.do(t, http.MethodPost, "/api/libs", map[string]any{"libName": "方案源库"}, env.token)
	libId := res.num(t, "id")
	env.do(t, http.MethodPost, fmt.Sprintf("/api/libs/%d/words", libId), map[string]any{"word": "apple"}, env.token)

	// 单词为空 → 400
	res = env.do(t, http.MethodPost, "/api/plans", map[string]any{"name": "空方案", "libId": libId}, env.token)
	if res.status != 400 {
		t.Errorf("空方案状态码 = %d, 期望 400", res.status)
	}

	res = env.do(t, http.MethodPost, "/api/plans", map[string]any{
		"name": "期末听写", "libId": libId, "disorder": true, "ch2en": false, "repeat": 2,
		"words": []map[string]string{{"word": "apple", "explain": "苹果"}},
	}, env.token)
	if res.status != 200 {
		t.Fatalf("创建方案失败: %s", res.body)
	}
	planId := res.num(t, "id")

	res = env.do(t, http.MethodGet, "/api/plans", nil, env.token)
	if res.status != 200 || !strings.Contains(res.body, `"cnt":1`) || !strings.Contains(res.body, "期末听写") {
		t.Errorf("方案列表异常: %s", res.body)
	}

	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/plans/%d/words", planId), nil, env.token)
	if res.status != 200 || !strings.Contains(res.body, "apple") {
		t.Errorf("方案单词异常: %s", res.body)
	}

	res = env.do(t, http.MethodDelete, fmt.Sprintf("/api/plans/%d", planId), nil, env.token)
	if res.status != 200 {
		t.Fatalf("删除方案失败: %s", res.body)
	}
	res = env.do(t, http.MethodGet, fmt.Sprintf("/api/plans/%d/words", planId), nil, env.token)
	if res.status != 404 {
		t.Errorf("删除后访问方案状态码 = %d, 期望 404", res.status)
	}
}
