package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func getPolLang() (LM, error) {
	path := "data/polish_lang.csv"
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(fc))
	r.Comma = rune(',')
	r.Comment = rune('#')
	var ret = make(LM)
	var rec []string
	rec, err = r.Read()
	for err == nil {
		code := rec[0]
		pl := rec[1]

		ret[code] = Lang{
			Pl: pl,
		}

		rec, err = r.Read()
	}

	if err == io.EOF {
		err = nil
	}

	fmt.Printf("Got %d langs from %s\n", len(ret), path)
	return ret, err
}

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
			Eng:       p.Data[l],
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

	plLM, err := getPolLang()
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
			rl.Eng = gl.Eng
			rl.ISO_639_2 = gl.ISO_639_2
			rl.Fr = gl.Fr
			rl.Ger = gl.Ger
		}

		if wl, f := wikiLM[l]; !f {
			continue
		} else {
			rl.Family = wl.Family
			rl.Endonym = wl.Endonym
		}

		if pl, f := plLM[l]; !f {
			continue
		} else {
			rl.Pl = pl.Pl
		}

		joined[rl.ISO_639_1] = rl
	}

	fmt.Printf("%d langs after join\n", len(joined))

	return writeToJson("joined.json", joined)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
