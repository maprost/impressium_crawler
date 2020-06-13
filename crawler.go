package imprint_crawler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type MainPage struct {
	Link     string
	Err      error
	Imprints map[string]Imprint
}

func (p MainPage) String() string {
	s := fmt.Sprintln("Link: ", p.Link)
	s += fmt.Sprintln("Error: ", p.Err)
	s += fmt.Sprintln("Imprint:")
	for _, i := range p.Imprints {
		s += fmt.Sprintln(i)
	}

	return s
}

type Imprint struct {
	Tag  string
	Link string
	Err  error

	Name    string
	Address string
	Zip     string
	Email   string
}

func (i Imprint) String() string {
	s := fmt.Sprintln("\tTag/Link: ", i.Tag, i.Link)

	s += fmt.Sprintln("\tError: ", i.Err)

	s += fmt.Sprintln("\tName: ", i.Name)
	s += fmt.Sprintln("\tAddress: ", i.Address)
	s += fmt.Sprintln("\tZip: ", i.Zip)
	s += fmt.Sprintln("\tEmail: ", i.Email)

	return s
}

var possibleTags = []string{"Kontakt", "Impressum"}

func CrawlMainPages(links []string) []MainPage {
	result := make([]MainPage, 0, len(links))
	for _, link := range links {
		result = append(result, CrawlMainPage(link))
	}

	return result
}

func CrawlMainPage(link string) MainPage {
	mainPage := MainPage{
		Link:     link,
		Imprints: make(map[string]Imprint),
	}

	resp, err := http.Get(link)
	if err != nil {
		mainPage.Err = err
		return mainPage
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	for {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
			if len(mainPage.Imprints) == 0 {
				mainPage.Err = errors.New("can't finde imprint on mainpage")
			}
			return mainPage

		case tokenType == html.StartTagToken:
			token := z.Token()

			if token.Data == "a" {
				// need the '<a>..</a>' token value
				if z.Next() != html.TextToken {
					continue
				}
				tokenValue := z.Token()

				addImprint(mainPage, token, tokenValue)
			}
		}
	}
}

func addImprint(mainPage MainPage, token html.Token, tokenValue html.Token) {
	//fmt.Println("check:", tokenValue, token)
	imprint := Imprint{}

	// check if the <a> tag value is correct
	found, tag := checkTag(tokenValue)
	if !found {
		return
	}
	imprint.Tag = strings.Replace(tag, "\n", "", -1)

	// get link
	imprint.Link = strings.Replace(getHref(token), "\n", "", -1)
	if imprint.Link == "" {
		return
	}

	if strings.HasPrefix(imprint.Link, "http") == false {
		imprint.Link = mainPage.Link + imprint.Link
	}

	// crawl imprint
	crawlImprint(&imprint)
	mainPage.Imprints[imprint.Link] = imprint
}

func checkTag(tokenValue html.Token) (bool, string) {
	for _, tag := range possibleTags {
		s := tokenValue.String()
		if strings.Contains(s, tag) {
			return true, s
		}
	}

	return false, ""
}

func getHref(token html.Token) string {
	for _, a := range token.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}

	return ""
}

func crawlImprint(imprint *Imprint) {
	resp, err := http.Get(imprint.Link)
	if err != nil {
		imprint.Err = err
		return
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	zipMatch := false
	emailMatch := false
	lastValue := ""
	lastLastValue := ""

	for {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
			if zipMatch == false {
				imprint.Err = errors.New("address not found")
			}

			if emailMatch == false {
				imprint.Err = errors.New("email not found")
			}
			return

		case tokenType == html.TextToken:
			token := z.Token()
			value := token.String()

			// check zip code
			if zipMatch == false {
				zipMatch, _ = regexp.MatchString("[0-9][0-9][0-9][0-9][0-9] [A-Z][a-z]", value)
				if zipMatch {
					imprint.Name = lastLastValue
					imprint.Address = lastValue
					imprint.Zip = value
				}
			}

			// check email
			if emailMatch == false {
				emailMatch, _ = regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", value)
				if emailMatch {
					imprint.Email = value
				}
			}

			lastLastValue = lastValue
			lastValue = value

		}
	}
}
