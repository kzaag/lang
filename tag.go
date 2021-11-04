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

type Sport struct {
	Name string
	Href string
}

type categoryMap map[string][]Sport

var wc int64
var tagCount int64

func (cm categoryMap) Add(category, discipline, href string) {
	x := cm[category]
	if x == nil {
		x = make([]Sport, 0, 10)
	}
	x = append(x, Sport{
		Name: discipline,
		Href: href,
	})
	cm[category] = x
	wc += int64(len(discipline))
	tagCount++
}

func setSports(c *html.Node, category string, cm categoryMap) bool {
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

func genSports() error {
	url := "https://en.wikipedia.org/wiki/List_of_sports"
	doc, err := getHtmlFromRemote("cache/List_of_sports.html", url)
	if err != nil {
		return err
	}

	type cfn func(n *html.Node, cm categoryMap, category string)

	cm := make(categoryMap)

	var x cfn

	x = func(n *html.Node, cm categoryMap, category string) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if setCategory(c, &category) {
				// if category was set then dont descend
				continue
			}
			setSports(c, category, cm)
			x(c, cm, category)
		}
	}

	x(doc, cm, "")

	b, err := json.MarshalIndent(cm, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("got %d sport categories\n", tagCount)

	return ioutil.WriteFile("data/sports.json", b, 0600)
}
