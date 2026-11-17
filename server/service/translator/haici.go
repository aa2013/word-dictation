package translator

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HaiCi 海词 dict.cn，HTML 解析
type HaiCi struct{}

func (h *HaiCi) Name() string { return "haici" }

func (h *HaiCi) Translate(word string) (Result, error) {
	html, err := httpGet("https://dict.cn/search?q=" + encodeWord(word))
	if err != nil {
		return Result{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Result{}, err
	}
	if doc.Find(".ifufind").Length() > 0 {
		return Result{}, fmt.Errorf("未查询到 %q", word)
	}

	res := Result{}
	bdo := doc.Find(".phonetic bdo")
	if bdo.Length() > 0 {
		res.EnSymbol = trimBrackets(bdo.Eq(0).Text())
	}
	if bdo.Length() > 1 {
		res.UsSymbol = trimBrackets(bdo.Eq(1).Text())
	}

	fsounds := doc.Find(".phonetic .sound.fsound")
	if fsounds.Length() > 0 {
		if v, ok := fsounds.Eq(0).Attr("naudio"); ok {
			res.EnMp3 = "https://audio.dict.cn/" + v
		}
	}
	if fsounds.Length() > 1 {
		if v, ok := fsounds.Eq(1).Attr("naudio"); ok {
			res.UsMp3 = "https://audio.dict.cn/" + v
		}
	}

	lis := doc.Find(".basic.clearfix li")
	res.Explains = []Explain{}
	// 末尾一项为页面附加元素，与旧实现一致地跳过
	for i := 0; i < lis.Length()-1; i++ {
		li := lis.Eq(i)
		children := li.Children()
		if children.Length() < 2 {
			continue
		}
		typ := children.Eq(0).Text()
		for _, mean := range strings.Split(children.Eq(1).Text(), "；") {
			if mean != "" {
				res.Explains = append(res.Explains, Explain{Explanation: mean, Type: typ})
			}
		}
	}
	if len(res.Explains) == 0 {
		return Result{}, fmt.Errorf("未查询到 %q", word)
	}
	return res, nil
}

func trimBrackets(s string) string {
	s = strings.TrimSpace(s)
	return strings.NewReplacer("[", "", "]", "").Replace(s)
}
