package translator

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CiBa 爱词霸 iciba.com，解析页面 __NEXT_DATA__ JSON
type CiBa struct{}

func (c *CiBa) Name() string { return "ciba" }

type cibaNextData struct {
	Props struct {
		PageProps struct {
			InitialReduxState struct {
				Word struct {
					WordInfo struct {
						BaesInfo struct {
							Symbols []struct {
								PhEn    string `json:"ph_en"`
								PhAm    string `json:"ph_am"`
								PhEnMp3 string `json:"ph_en_mp3"`
								PhAmMp3 string `json:"ph_am_mp3"`
								Parts   []struct {
									Part  string   `json:"part"`
									Means []string `json:"means"`
								} `json:"parts"`
							} `json:"symbols"`
						} `json:"baesInfo"`
					} `json:"wordInfo"`
				} `json:"word"`
			} `json:"initialReduxState"`
		} `json:"pageProps"`
	} `json:"props"`
}

func (c *CiBa) Translate(word string) (Result, error) {
	html, err := httpGet("https://www.iciba.com/word?w=" + encodeWord(word))
	if err != nil {
		return Result{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Result{}, err
	}
	if doc.Find(".Mean_word__hwr_g").Length() == 0 {
		return Result{}, fmt.Errorf("未查询到 %q", word)
	}
	el := doc.Find("#__NEXT_DATA__")
	if el.Length() == 0 {
		return Result{}, fmt.Errorf("页面结构变化，无法解析")
	}
	var data cibaNextData
	if err := json.Unmarshal([]byte(el.Text()), &data); err != nil {
		return Result{}, err
	}
	symbols := data.Props.PageProps.InitialReduxState.Word.WordInfo.BaesInfo.Symbols
	if len(symbols) == 0 {
		return Result{}, fmt.Errorf("翻译结果为空")
	}
	sym := symbols[0]

	res := Result{
		EnSymbol: sym.PhEn,
		EnMp3:    sym.PhEnMp3,
		UsSymbol: sym.PhAm,
		UsMp3:    sym.PhAmMp3,
		Explains: []Explain{},
	}
	for _, part := range sym.Parts {
		for _, mean := range part.Means {
			res.Explains = append(res.Explains, Explain{Explanation: mean, Type: part.Part})
		}
	}
	sort.SliceStable(res.Explains, func(i, j int) bool {
		return len(res.Explains[i].Explanation) < len(res.Explains[j].Explanation)
	})
	return res, nil
}
