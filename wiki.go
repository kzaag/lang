package main

import (
	"golang.org/x/net/html"
)

// c must be TR node (c.Data must be "tr")
func parseWikiTr(c *html.Node, l *Lang) bool {

	// first tr child
	c = c.FirstChild
	if c == nil {
		return false
	}
	// first td
	c = c.NextSibling
	if c == nil {
		return false
	}

	// first td is empty, skipping to 2nd
	if c = skipSibling(c); c == nil || c.FirstChild.Data != "a" {
		return false
	}

	l.Family = c.FirstChild.FirstChild.Data

	if c = skipSibling(c); c == nil || c.FirstChild.Data != "a" {
		return false
	}
	l.Eng = c.FirstChild.FirstChild.Data

	if c = skipSibling(c); c == nil {
		return false
	}

	// sometimes endonym is plaintext, sometimes its html.
	if c.FirstChild.Data == "div" || c.FirstChild.Data == "a" || c.FirstChild.Data == "i" || c.FirstChild.Data == "span" {
		l.Endonym = c.FirstChild.FirstChild.Data
	} else {
		l.Endonym = c.FirstChild.Data
	}

	if c = skipSibling(c); c == nil {
		return false
	}

	// sure, why not
	l.ISO_639_1 = c.FirstChild.FirstChild.FirstChild.FirstChild.Data

	return true
}

func readWikiNode(n *html.Node, langs LM) {
	recReadTds(n, langs, parseWikiTr)
}

func getLangFromWiki() (LM, error) {
	return crawlRemote(
		"cache/en.wikipedia.org.html",
		"https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes",
		readWikiNode)
}
