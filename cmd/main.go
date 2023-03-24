package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ibrahimduran/answer-cli/pkg/ansifyhtml"
	"github.com/ibrahimduran/answer-cli/pkg/answerlib"
	term "github.com/nsf/termbox-go"
	"github.com/pterm/pterm"
)

var sources = []answerlib.Source{
	{
		Name:      "StackOverflow",
		Patterns:  []string{"stackoverflow\\.com/questions/[0-9]+"},
		Extractor: "#answers .post-layout .js-post-body",
	},
	{
		Name:      "StackAcademia",
		Patterns:  []string{"academia.stackexchange\\.com/questions/[0-9]+"},
		Extractor: "#answers .post-layout .js-post-body",
	},
}

func init() {
	flag.Parse()
}

func main() {
	query := strings.Join(os.Args[1:], " ")

	search_url := url.URL{
		Scheme:   "https",
		Host:     "google.com",
		Path:     "search",
		RawQuery: fmt.Sprintf("q=%s", url.QueryEscape(query)),
	}

	res, err := http.Get(search_url.String())

	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	answers := []string{}

	doc.Find("a[href]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		val, exists := s.Attr("href")

		if !strings.HasPrefix(val, "http") {
			parsed, err := url.Parse(val)
			if err != nil {
				panic(err)
			}

			if parsed.Host == "" {
				parsed.Host = search_url.Host
			}

			if parsed.Scheme == "" {
				parsed.Scheme = search_url.Scheme
			}

			val = parsed.String()
		}

		if !exists {
			return true
		}

		for _, source := range sources {
			for _, pattern := range source.Patterns {
				if matched, err := regexp.MatchString(pattern, val); err != nil {
					log.Fatal(err)
				} else if matched {
					answers = append(answers, getAnswers(val, &source)...)
					return false
				}
			}
		}

		return true
	})

	var content string

	if len(answers) == 0 {
		content = "Nothing found :("
		pterm.DefaultBox.WithTitle("Lorem Ipsum").WithTitleBottomCenter().WithRightPadding(0).WithBottomPadding(0).Println(content)
	}

	if err := term.Init(); err != nil {
		panic(err)
	}

	defer term.Close()

	active := 0

	area, _ := pterm.DefaultArea.WithCenter().Start()
	defer area.Stop()

	for {
		content = answers[active]
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = pterm.DefaultParagraph.WithMaxWidth(pterm.GetTerminalWidth() - 10).Sprintln(line)
		}
		p := strings.Join(lines, "\n")
		box := pterm.DefaultBox.WithTitle(fmt.Sprintf("%d / %d", active+1, len(answers))).WithTitleBottomCenter().WithRightPadding(0).WithBottomPadding(0).Sprint(p)
		area.Update(box)

		evt := term.PollEvent()

		if evt.Key == term.KeyArrowLeft {
			if active != 0 {
				active -= 1
			}
		} else if evt.Key == term.KeyArrowRight {
			if active != len(answers)-1 {
				active += 1
			}
		} else {
			break
		}
	}
}

func getAnswers(url string, source *answerlib.Source) []string {
	res, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	results := []string{}

	doc.Find(source.Extractor).Each(func(i int, s *goquery.Selection) {
		text := ansifyhtml.Ansify(s.Nodes)
		space := regexp.MustCompile(`[ ]+`)
		text = space.ReplaceAllString(text, " ")
		space2 := regexp.MustCompile(`\n+`)
		text = space2.ReplaceAllString(text, "\n")
		results = append(results, res.Request.URL.String()+"\n"+text)
	})

	return results
}
