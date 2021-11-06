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

func persistData(path string, d interface{}) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, b, 0600)
}

type TranslatedCategory struct {
	Name         string
	Translations map[string]string
}

type IsoLang struct {
	ISO_639_1 string
}

func translateCategories(
	cm CategoryMap,
	langs []IsoLang,
	supportedLangMap map[string]struct{},
	token *GoogleToken) ([]TranslatedCategory, error) {

	var res = make([]TranslatedCategory, len(cm))
	var cstr = make([]string, len(res))
	var i = 0
	for c := range cm {
		cstr[i] = c
		res[i] = TranslatedCategory{
			Name:         c,
			Translations: make(map[string]string),
		}
		i++
	}
	var tres TranslateResponse

	fmt.Println("translating categories")

	for i := range langs {
		l := langs[i].ISO_639_1
		if _, e := supportedLangMap[l]; !e {
			continue
		}
		if l == "en" {
			continue
		}
		if err := GoogleTranslate(token, &TranslateRequest{
			Q:      cstr,
			Source: "en",
			Target: l,
			Format: "text",
		}, &tres); err != nil {
			fmt.Printf("Couldnt translate to %s, reason: %v\n", l, err)
			break
		}
		if len(tres.Data.Translations) != len(res) {
			fmt.Printf("Couldnt translate to %s, reason: invalid translation length\n", l)
			break
		}
		for z := range tres.Data.Translations {
			res[z].Translations[l] = tres.Data.Translations[z].TranslatedText
		}
	}

	return res, nil
}

func translateDisciplines(
	cm CategoryMap,
	langs []IsoLang,
	supportedLangMap map[string]struct{},
	token *GoogleToken,
) ([]TranslatedSportDiscipline, error) {

	l := 0
	for c := range cm {
		l += len(cm[c])
	}

	const savePointPath = "data/continue.sports.json"

	res := make([]TranslatedSportDiscipline, l)
	enDisciplines := make([]string, l)
	translatedLangs := make(map[string]int)
	// read continuation point
	cfc, err := ioutil.ReadFile(savePointPath)
	if err == nil {
		var cres []TranslatedSportDiscipline
		if err := json.Unmarshal(cfc, &cres); err != nil {
			return nil, err
		}
		if len(cres) != l {
			return nil, fmt.Errorf("cannot resume translating - invalid length")
		}
		for i := range cres {
			enDisciplines[i] = cres[i].Name
			for l := range cres[i].Translations {
				if cres[i].Translations[l] != "" {
					translatedLangs[l]++
				}
			}
		}
		res = cres

		for lang := range translatedLangs {
			if translatedLangs[lang] != l {
				return nil, fmt.Errorf("language: %s seem to be missing some translations", lang)
			}
		}

	} else {
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
	}

	const batchSize = 100
	var tres TranslateResponse

	var isConfirmed bool

	// translate names
	for i := range langs {
		dstLang := langs[i].ISO_639_1
		if translatedLangs[dstLang] != 0 || dstLang == "en" {
			fmt.Println("already translated " + dstLang)
			continue
		}
		if _, e := supportedLangMap[dstLang]; !e {
			fmt.Printf("lang %s is not supported\n", dstLang)
			continue
		}

		fmt.Printf("translating to %s\n", dstLang)

		if !isConfirmed {
			fmt.Println("About to perform google translations, are you sure? [yes/no]")
			var tmp string
			fmt.Scanln(&tmp)
			if tmp == "yes" {
				isConfirmed = true
			} else {
				return nil, fmt.Errorf("aborted by user")
			}
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

		// save progress
		if err := persistData(savePointPath, res); err != nil {
			return nil, err
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

	fmt.Printf("got %d sport disciplines\n", tagCount)

	// read langs
	fc, err := ioutil.ReadFile("data/joined.json")
	if err != nil {
		return err
	}

	var langs []IsoLang

	if err := json.Unmarshal(fc, &langs); err != nil {
		return err
	}

	token, err := NewGoogleTranslateToken()
	if err != nil {
		return err
	}
	token.AccessToken = strings.TrimRight(token.AccessToken, ".")

	var tl DiscoverLangResponse
	if err := GoogleDiscoverLangs(token, &tl); err != nil {
		return err
	}
	supportedLangMap := make(map[string]struct{})
	for i := range tl.Data.Languages {
		supportedLangMap[tl.Data.Languages[i].Language] = struct{}{}
	}

	ts, err := translateDisciplines(sportCategories, langs, supportedLangMap, token)
	if err != nil {
		return err
	}

	if err := persistData("data/sports.json", ts); err != nil {
		return err
	}

	tcat, err := translateCategories(sportCategories, langs, supportedLangMap, token)
	if err != nil {
		return err
	}

	return persistData("data/categories.json", tcat)
}
