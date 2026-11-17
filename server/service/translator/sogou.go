package translator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// SouGou 搜狗翻译，解析页面内 __INITIAL_STATE__ JSON
type SouGou struct{}

func (s *SouGou) Name() string { return "sogou" }

type sogouState struct {
	Voice struct {
		From struct {
			List []sogouPhonetic `json:"list"`
		} `json:"from"`
		Form *struct {
			List []sogouPhonetic `json:"list"`
		} `json:"form"`
	} `json:"voice"`
	TextTranslate struct {
		TranslateData struct {
			WordCard struct {
				UsualDict []struct {
					Pos    string   `json:"pos"`
					Values []string `json:"values"`
				} `json:"usualDict"`
			} `json:"wordCard"`
		} `json:"translateData"`
	} `json:"textTranslate"`
}

type sogouPhonetic struct {
	Type     string `json:"type"`
	Phonetic string `json:"phonetic"`
	Src      string `json:"src"`
}

const sogouStatePrefix = "window.__INITIAL_STATE__="
const sogouStateSuffix = ";(function(){var s;(s=document.currentScript||document.scripts[document.scripts.length-1]).parentNode.removeChild(s);}());"

func (s *SouGou) Translate(word string) (Result, error) {
	html, err := httpGet("https://fanyi.sogou.com/text?keyword=" + encodeWord(word))
	if err != nil {
		return Result{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Result{}, err
	}
	idx := doc.Find("#translateIndex")
	if idx.Length() == 0 {
		return Result{}, fmt.Errorf("未查询到 %q", word)
	}
	script := strings.TrimSpace(idx.Next().Text())
	script = strings.TrimPrefix(script, sogouStatePrefix)
	script = strings.TrimSuffix(script, sogouStateSuffix)
	var state sogouState
	if err := json.Unmarshal([]byte(script), &state); err != nil {
		return Result{}, err
	}

	res := Result{Explains: []Explain{}}
	for _, p := range state.Voice.From.List {
		if strings.Contains(p.Type, "uk") {
			res.EnSymbol = p.Phonetic
		} else if strings.Contains(p.Type, "us") {
			res.UsSymbol = p.Phonetic
		}
	}
	if state.Voice.Form != nil {
		for _, p := range state.Voice.Form.List {
			if strings.Contains(p.Type, "uk") {
				res.EnMp3 = "https:" + p.Src
			} else if strings.Contains(p.Type, "us") {
				res.UsMp3 = "https:" + p.Src
			}
		}
	}
	usual := state.TextTranslate.TranslateData.WordCard.UsualDict
	for _, dict := range usual {
		for _, v := range dict.Values {
			for _, mean := range strings.Split(v, "；") {
				if mean != "" {
					res.Explains = append(res.Explains, Explain{Explanation: mean, Type: dict.Pos})
				}
			}
		}
	}
	if len(res.Explains) == 0 {
		return Result{}, fmt.Errorf("翻译结果为空")
	}
	return res, nil
}
