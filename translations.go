package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func getLocalName(code string) (map[string]string, error) {
	path := fmt.Sprintf("language-list/data/%s/language.csv", code)
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(fc))
	r.Comma = rune(',')
	r.Comment = rune('#')
	var ret = make(map[string]string)
	var rec []string
	rec, err = r.Read()
	for err == nil {
		code := rec[0]
		n := rec[1]
		ret[code] = n
		rec, err = r.Read()
	}

	if err == io.EOF {
		err = nil
	}

	return ret, err
}

type Translations map[string]map[string]string

func newTranslations(lm LM) (Translations, error) {
	pivot := make(Translations)
	for l := range lm {
		m, err := getLocalName(l)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		pivot[l] = m
	}

	ret := make(Translations, len(pivot))
	for l := range pivot {
		for t := range pivot[l] {
			if ret[t] == nil {
				ret[t] = make(map[string]string)
			}
			ret[t][l] = pivot[l][t]
		}
	}

	return ret, nil
}
