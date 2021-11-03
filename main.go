package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func getLangFromJsonLM() (LM, error) {
	path := "data/iso639-1_loc.json"
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p struct {
		Data map[string]string
	}
	if err := json.Unmarshal(fc, &p); err != nil {
		return nil, err
	}
	var ret = make(LM)
	for l := range p.Data {
		if len(l) != 2 {
			continue // ignoring locale
		}
		ret[l] = Lang{
			ISO_639_1: l,
			En:        p.Data[l],
		}
	}
	fmt.Printf("Got %d langs from %s\n", len(ret), path)
	return ret, nil
}

func run() error {

	govLM, err := getLangFromGov()
	if err != nil {
		return err
	}

	wikiLM, err := getLangFromWiki()
	if err != nil {
		return err
	}

	jsonLM, err := getLangFromJsonLM()
	if err != nil {
		return err
	}

	translations, err := newTranslations(wikiLM)
	if err != nil {
		return err
	}

	var joined = make(LM)

	for l := range jsonLM {
		var rl = Lang{
			ISO_639_1: l,
		}
		if gl, f := govLM[l]; !f {
			continue
		} else {
			rl.En = strings.ToLower(gl.En)
			rl.ISO_639_2 = strings.ToLower(gl.ISO_639_2)
			rl.Fr = strings.ToLower(gl.Fr)
			rl.De = strings.ToLower(gl.De)
		}

		if wl, f := wikiLM[l]; !f {
			continue
		} else {
			rl.Family = strings.ToLower(wl.Family)
			rl.Endonym = strings.ToLower(wl.Endonym)
		}

		if translations[l] == nil {
			continue
		}

		rl.Translations = translations[l]

		joined[strings.ToLower(rl.ISO_639_1)] = rl
	}

	fmt.Printf("%d langs after join\n", len(joined))

	return writeToJson("data/joined.json", joined)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
