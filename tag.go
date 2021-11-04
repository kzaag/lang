package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
)

func setCategory(c *html.Node, category *string) bool {

	if c == nil || (c.Data != "h3" && c.Data != "h2") {
		return false
	}
	cc := c.FirstChild
	if cc == nil || cc.Data != "span" {
		return false
	}
	for i := range cc.Attr {
		if cc.Attr[i].Key == "class" && cc.Attr[i].Val == "mw-headline" {
			_category := strings.TrimSpace(cc.FirstChild.Data)
			if _category != "" {
				*category = _category
				wc += 1
				return true
			}
		}
	}
	return false
}

type SportDiscipline struct {
	Name string
	Href string
}

type TranslatedSportDiscipline struct {
	SportDiscipline
	Category     string
	Translations map[string]string
}

type CategoryMap map[string][]SportDiscipline

var wc int64
var tagCount int64

func (cm CategoryMap) Add(category, discipline, href string) {
	x := cm[category]
	if x == nil {
		x = make([]SportDiscipline, 0, 10)
	}
	x = append(x, SportDiscipline{
		Name: discipline,
		Href: href,
	})
	cm[category] = x
	wc += int64(len(discipline))
	tagCount++
}

func setSports(c *html.Node, category string, cm CategoryMap) bool {
	if category == "" {
		return false
	}
	if c == nil || c.Data != "li" {
		return false
	}
	cc := c.FirstChild
	if cc == nil || cc.Data != "a" {
		// no anchor but still may be discipline
		maybeDiscipline := strings.TrimSpace(cc.Data)
		if maybeDiscipline == "" {
			return false
		}
		cm.Add(category, maybeDiscipline, "")
		return true
	}

	href := ""
	for i := range cc.Attr {
		if cc.Attr[i].Key == "href" {
			href = cc.Attr[i].Val
			break
		}
	}

	if href == "" {
		return false
	}

	cm.Add(category, cc.FirstChild.Data, href)

	return true
}

func persistSportDisciplines(path string, sm []TranslatedSportDiscipline) error {
	b, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, b, 0600)
}

func translateCategories(cm CategoryMap) ([]TranslatedSportDiscipline, error) {

	// read langs
	fc, err := ioutil.ReadFile("data/joined.json")
	if err != nil {
		return nil, err
	}

	var langs []struct {
		ISO_639_1 string
	}

	if err := json.Unmarshal(fc, &langs); err != nil {
		return nil, err
	}

	l := 0
	for c := range cm {
		l += len(cm[c])
	}

	res := make([]TranslatedSportDiscipline, l)

	enDisciplines := make([]string, l)
	var i int
	for category := range cm {
		disciplines := cm[category]
		for j := range disciplines {
			enDisciplines[i] = disciplines[j].Name
			res[i].Category = category
			res[i].SportDiscipline = disciplines[j]
			res[i].Translations = make(map[string]string)
			i++
		}
	}

	if i != l {
		panic("invalid category count")
	}

	token, err := NewGoogleTranslateToken()
	if err != nil {
		return nil, err
	}
	var tres TranslateResponse
	token.AccessToken = strings.TrimRight(token.AccessToken, ".")
	const batchSize = 100

	for i := range langs {
		dstLang := langs[i].ISO_639_1
		if dstLang != "pl" {
			continue
		}

		offset := 0

		for {
			if offset == -1 {
				break
			}
			left := offset
			right := offset + batchSize

			// last batch
			if right >= len(enDisciplines) {
				right = len(enDisciplines)
				offset = -1
			} else {
				offset += batchSize
			}

			if right == left {
				break
			}
			batch := enDisciplines[left:right]

			if err := GoogleTranslate(token, &TranslateRequest{
				Q:      batch,
				Source: "en",
				Target: dstLang,
				Format: "text",
			}, &tres); err != nil {
				fmt.Printf("Couldnt translate to %s, reason: %v\n", dstLang, err)
				break
			}
			if len(tres.Data.Translations) != len(batch) {
				fmt.Printf("Couldnt translate to %s, reason: invalid translation length\n", dstLang)
				break
			}
			for z := range tres.Data.Translations {
				res[left+z].Translations[dstLang] = tres.Data.Translations[z].TranslatedText
			}
		}
	}

	return res, nil
}

func genSports() error {
	url := "https://en.wikipedia.org/wiki/List_of_sports"
	doc, err := getHtmlFromRemote("cache/List_of_sports.html", url)
	if err != nil {
		return err
	}

	type cfn func(n *html.Node, cm CategoryMap, category string)

	sportCategories := make(CategoryMap)

	var x cfn

	x = func(n *html.Node, cm CategoryMap, category string) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if setCategory(c, &category) {
				// if category was set then dont descend
				continue
			}
			setSports(c, category, cm)
			x(c, cm, category)
		}
	}

	x(doc, sportCategories, "")

	fmt.Printf("got %d sport categories\n", tagCount)

	ts, err := translateCategories(sportCategories)
	if err != nil {
		return err
	}

	return persistSportDisciplines("data/sports.json", ts)
}
