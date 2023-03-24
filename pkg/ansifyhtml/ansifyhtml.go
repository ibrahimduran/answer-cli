package ansifyhtml

import (
	"regexp"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/net/html"
)

func Ansify(nodes []*html.Node) string {
	str := ""

	for _, node := range nodes {
		str += traverse(node)
	}

	return strings.Trim(str, " \n")
}

func traverse(n *html.Node) string {
	str := ""
	var cb func(string) string

	switch n.Type {
	case html.TextNode:
		str += regexp.MustCompile(`\s+`).ReplaceAllString(n.Data, " ")
	case html.ElementNode:
		switch n.Data {
		case "p":
			str += "\n\n"
		case "code":
			cb = func(s string) string {
				return pterm.BgWhite.Sprint(pterm.FgBlack.Sprint(s))
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if cb == nil {
			str += traverse(c)
		} else {
			str += cb(traverse(c))
		}
	}

	return str
}
