package main

import (
	"strings"
)

func langsToLower(lm LM) LM {
	cpy := make(LM)
	for l := range lm {
		x := lm[l]
		xcpy := Lang{
			Endonym:      strings.ToLower(x.Endonym),
			En:           strings.ToLower(x.En),
			De:           strings.ToLower(x.De),
			Fr:           strings.ToLower(x.Fr),
			ISO_639_1:    strings.ToLower(x.ISO_639_1),
			ISO_639_2:    strings.ToLower(x.ISO_639_2),
			Family:       strings.ToLower(x.Family),
			Translations: make(map[string]string),
		}
		for t := range x.Translations {
			xcpy.Translations[strings.ToLower(t)] = strings.ToLower(x.Translations[t])
		}
		cpy[strings.ToLower(l)] = xcpy
	}
	return cpy
}

func categoriesToLower(tcat []TranslatedCategory) []TranslatedCategory {
	cpy := make([]TranslatedCategory, len(tcat))
	for i := range tcat {
		cpy[i] = TranslatedCategory{
			Name:         strings.ToLower(tcat[i].Name),
			Translations: make(map[string]string),
		}
		for l := range tcat[i].Translations {
			x := tcat[i].Translations[l]
			cpy[i].Translations[strings.ToLower(l)] = strings.ToLower(x)
		}
	}
	return cpy
}

func disciplinesToLower(td []TranslatedSportDiscipline) []TranslatedSportDiscipline {
	cpy := make([]TranslatedSportDiscipline, len(td))
	for i := range td {
		cpy[i] = TranslatedSportDiscipline{
			SportDiscipline: SportDiscipline{
				Name: strings.ToLower(td[i].SportDiscipline.Name),
				Href: td[i].SportDiscipline.Href,
			},
			Category:     strings.ToLower(td[i].Category),
			Translations: make(map[string]string),
		}
		for l := range td[i].Translations {
			x := td[i].Translations[l]
			cpy[i].Translations[strings.ToLower(l)] = strings.ToLower(x)
		}
	}
	return cpy
}
