package main

import (
	"strings"

	"golang.org/x/net/html"
)

// c must be TR node (c.Data must be "tr")
func parseGovTr(c *html.Node, l *Lang) bool {

	c = c.FirstChild
	if c == nil {
		return false
	}
	c = c.NextSibling
	if c == nil {
		return false
	}

	// first cell
	l.ISO_639_2 = strings.TrimSpace(c.FirstChild.Data)

	// for some reason this table has empty elements between td-s.
	// im just skipping those
	if c = skipSibling(c); c == nil {
		return false
	}

	// 2nd cell
	l.ISO_639_1 = strings.TrimSpace(c.FirstChild.Data)

	if c = skipSibling(c); c == nil {
		return false
	}

	// 3rd cell
	l.Eng = strings.TrimSpace(c.FirstChild.Data)

	if c = skipSibling(c); c == nil {
		return false
	}

	// 4th cell
	l.Fr = strings.TrimSpace(c.FirstChild.Data)

	if c = skipSibling(c); c == nil {
		return false
	}

	// 5th cell
	l.Ger = strings.TrimSpace(c.FirstChild.Data)

	return true
}

func readGovNode(n *html.Node, langs LM) {
	recReadTds(n, langs, parseGovTr)
}

func getLangFromGov() (LM, error) {
	return crawlRemote("cache/www.loc.gov.html", "https://www.loc.gov/standards/iso639-2/php/code_list.php", readGovNode)
}
