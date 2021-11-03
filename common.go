package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

type Lang struct {
	Endonym string
	// you can also find en de and fr in Translations map
	// however translations presented here are taken from different
	// - more reliable source and thus should be prioritized
	En string
	De string
	Fr string
	//
	ISO_639_1    string
	ISO_639_2    string
	Family       string
	Translations map[string]string
}

// ISO 639-1 is the key
type LM map[string]Lang

func skipSibling(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	if n.NextSibling == nil {
		return nil
	}
	ret := n.NextSibling.NextSibling
	if ret == nil || ret.FirstChild == nil {
		return nil
	}
	return ret
}

func recReadTds(n *html.Node, langs LM, trCallback func(*html.Node, *Lang) bool) {
	var lang Lang
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "tr" {
			if ok := trCallback(c, &lang); ok {
				langs[lang.ISO_639_1] = lang
			}
		}
		recReadTds(c, langs, trCallback)
	}
}

func crawlRemote(fn string, _url string, cf func(n *html.Node, langs LM)) (LM, error) {

	rurl, err := url.Parse(_url)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(fn)
	var buff []byte
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		fmt.Printf("warn: cache not found (%s). GETting from %s\n", fn, _url)
		//url := "https://www.loc.gov/standards/iso639-2/php/code_list.php"
		req, err := http.NewRequest("GET", _url, nil)
		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf(res.Status)
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(fn, b, 0600); err != nil {
			return nil, err
		}
		buff = b
	} else {
		buff, err = ioutil.ReadFile(fn)
		if err != nil {
			return nil, err
		}
	}

	// TODO: should get encoding from meta
	e, name, _ := charset.DetermineEncoding(buff, "")
	if name != "utf-8" {
		fmt.Printf("warn: noticed encoding: %s. will convert.\n", name)
		r := transform.NewReader(bytes.NewBuffer(buff), e.NewDecoder())
		buff, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
	}

	doc, err := html.Parse(bytes.NewBuffer(buff))
	if err != nil {
		return nil, err
	}

	langs := make(LM)

	cf(doc, langs)

	fmt.Printf("Got %d langs from %s\n", len(langs), rurl.Host)

	return langs, nil
}

func writeToJson(p string, lm LM) error {
	ll := make([]Lang, 0, len(lm))
	i := 0
	for l := range lm {
		if l == "" {
			continue
		}
		ll = append(ll, lm[l])
		i++
	}

	b, err := json.MarshalIndent(ll, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p, b, 0600)
}
