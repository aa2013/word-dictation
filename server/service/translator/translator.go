// Package translator 在线词典查询：海词 → 爱词霸 → 搜狗 依次降级。
package translator

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Explain struct {
	Explanation string
	Type        string
}

type Result struct {
	EnSymbol string
	EnMp3    string
	UsSymbol string
	UsMp3    string
	Explains []Explain
}

type Translator interface {
	Name() string
	Translate(word string) (Result, error)
}

var chain = []Translator{&HaiCi{}, &CiBa{}, &SouGou{}}

// SetTranslators 供测试替换翻译链
func SetTranslators(ts []Translator) { chain = ts }

func Translate(word string) (Result, error) {
	var lastErr error
	for _, t := range chain {
		res, err := t.Translate(word)
		if err == nil && len(res.Explains) > 0 {
			return res, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New("翻译结果为空")
		}
		log.Printf("[translator] %s 查询 %q 失败: %v", t.Name(), word, lastErr)
	}
	if lastErr == nil {
		lastErr = errors.New("单词不存在或查询源不可用")
	}
	return Result{}, lastErr
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func httpGet(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("查询源返回 " + resp.Status)
	}
	return string(body), nil
}

func encodeWord(word string) string {
	return url.QueryEscape(word)
}
