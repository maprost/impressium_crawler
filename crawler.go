package imprint_crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"golang.org/x/net/html"
)

const (
	LinkRedirectDiffers = "LinkRedirectDiffers"
)

type MainPage struct {
	Link     string
	Redirect string
	MainUrl  string
	Title    string
	Err      error
	Flag     string

	Address     string
	Zip         string
	City        string
	Email       string
	BestImprint *Imprint

	Imprints map[string]*Imprint
	Contacts map[string]*Imprint
}

func (p MainPage) String() string {
	s := fmt.Sprintln("Link: ", p.Link)
	s += fmt.Sprintln("Flag: ", p.Flag)
	s += fmt.Sprintln("Title: ", p.Title)
	s += fmt.Sprintln("Error: ", p.Err)
	s += fmt.Sprintln("Address: ", p.Address)
	s += fmt.Sprintln("Zip: ", p.Zip)
	s += fmt.Sprintln("City: ", p.City)
	s += fmt.Sprintln("Email: ", p.Email)
	s += fmt.Sprintln("BestImprint: ", p.BestImprint)
	s += fmt.Sprintln("Imprint:")
	for _, i := range p.Imprints {
		s += fmt.Sprintln(i)
	}
	s += fmt.Sprintln("Contact:")
	for _, i := range p.Contacts {
		s += fmt.Sprintln(i)
	}

	return s
}

func CSVHeader() string {
	return fmt.Sprintln("Link, Redirect, MainUrl, Title, Address, Zip, City, Email", ImprintCSVHeader())
}

func (p MainPage) CSV() string {
	return fmt.Sprintln(p.Link, " ,", p.Redirect, " ,", p.MainUrl, " ,", p.Title, " ,", p.Address, " ,", p.Zip, " ,", p.City, " ,", p.Email, p.BestImprint.CSV())
}

type Imprint struct {
	Tag  string
	Link string
	Err  error

	Name    string
	Address string
	Zip     string
	City    string
	Email   string
}

func (i Imprint) String() string {
	s := fmt.Sprintln("\tTag/Link: ", i.Tag, i.Link)

	s += fmt.Sprintln("\tError: ", i.Err)

	s += fmt.Sprintln("\tName: ", i.Name)
	s += fmt.Sprintln("\tAddress: ", i.Address)
	s += fmt.Sprintln("\tZip: ", i.Zip)
	s += fmt.Sprintln("\tCity: ", i.City)
	s += fmt.Sprintln("\tEmail: ", i.Email)

	return s
}

func ImprintCSVHeader() string {
	return fmt.Sprintln(", Imprint-Address, Imprint-Zip, Imprint-City, Imprint-Email")
}

func (i Imprint) CSV() string {
	return fmt.Sprint(" ,", i.Address, " ,", i.Zip, " ,", i.City, " ,", i.Email)
}

func CrawlMainPages(links []string) *Cache {
	cache := NewCache()
	cacheMutex := sync.Mutex{}

	pool := tunny.NewFunc(20, func(payload interface{}) interface{} {
		var result []byte
		link := payload.(string)

		cacheMutex.Lock()
		cache.MainPages[link] = CrawlMainPage(link)
		cacheMutex.Unlock()

		return result
	})
	defer pool.Close()

	start := time.Now()
	for _, link := range links {
		pool.Process(link)
	}
	fmt.Println("Process time:", time.Since(start))

	return cache
}

func CrawlMainPage(link string) MainPage {
	mainPage := MainPage{
		Link:        link,
		BestImprint: &Imprint{},
		Imprints:    make(map[string]*Imprint),
		Contacts:    make(map[string]*Imprint),
	}

	resp, err := http.Get(link)
	if err != nil {
		mainPage.Err = err
		return mainPage
	}

	mainPage.Redirect = resp.Request.URL.String()
	if mainPage.Redirect != mainPage.Link {
		r := strings.Replace(mainPage.Redirect, "http://", "", 1)
		r = strings.Replace(r, "http://", "", 1)

		l := strings.Replace(mainPage.Link, "http://", "", 1)
		l = strings.Replace(l, "http://", "", 1)

		if r != l {
			mainPage.Flag = LinkRedirectDiffers
		}
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	firstTitle := false
	zipMatch := false
	emailMatch := false
	lastValue := ""

	loop := true
	for loop {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
			loop = false

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

			if firstTitle == false && strings.ToLower(token.Data) == "title" {
				// need the '<title>..</title>' token value
				if z.Next() != html.TextToken {
					continue
				}
				tokenValue := z.Token()
				mainPage.Title = trim(tokenValue.String())
				firstTitle = true
			}

		case tokenType == html.TextToken:
			token := z.Token()
			value := token.String()

			// check zip code
			if zipMatch == false {
				zipMatch, _ = regexp.MatchString("[0-9][0-9][0-9][0-9][0-9] [A-Z][a-z]", value)
				if zipMatch {
					mainPage.Address = trim(lastValue)
					mainPage.Zip, mainPage.City = ZipCityTrimmer(value)
				}
			}

			// check email
			if emailMatch == false {
				emailMatch, _ = regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", value)
				if emailMatch {
					mainPage.Email = trim(value)
				}
			}

			lastValue = value
		}
	}

	if len(mainPage.Imprints) == 0 && len(mainPage.Contacts) == 0 {
		mainPage.Err = errors.New("can't finde imprint on mainpage")
	} else {
		chooseBestImprint(&mainPage)
	}
	return mainPage
}

func addImprint(mainPage MainPage, token html.Token, tokenValue html.Token) {
	//fmt.Println("check:", tokenValue, token)
	imprint := &Imprint{}

	// check if the <a> tag value is correct
	foundImprint, tag := isImprint(tokenValue)
	var foundContact bool
	if !foundImprint {
		foundContact, tag = isContact(tokenValue)
		if !foundContact {
			return
		}
	}
	imprint.Tag = strings.Replace(tag, "\n", "", -1)

	// get link
	imprint.Link = strings.Replace(getHrefValue(token), "\n", "", -1)
	if imprint.Link == "" {
		return
	}

	if strings.HasPrefix(imprint.Link, "http") == false {
		imprint.Link = concatLink(mainPage.Redirect, imprint.Link)
	}

	// crawl imprint
	crawlImprint(imprint)
	if foundImprint {
		mainPage.Imprints[imprint.Link] = imprint
	}
	if foundContact {
		mainPage.Contacts[imprint.Link] = imprint
	}
}

func isImprint(tokenValue html.Token) (bool, string) {
	s := tokenValue.String()
	if strings.Contains(strings.ToLower(s), "impressum") {
		return true, s
	}

	return false, ""
}

func isContact(tokenValue html.Token) (bool, string) {
	s := tokenValue.String()
	if strings.Contains(strings.ToLower(s), "kontakt") {
		return true, s
	}

	return false, ""
}

func getHrefValue(token html.Token) string {
	for _, a := range token.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}

	return ""
}

var client = http.Client{
	Timeout: 5 * time.Second,
}

func crawlImprint(imprint *Imprint) {
	// first try https
	imprint.Link = strings.Replace(imprint.Link, "http://", "https://", 1)

	resp, err := client.Get(imprint.Link)
	if err != nil {
		// second try http
		imprint.Link = strings.Replace(imprint.Link, "https://", "http://", 1)

		resp, err = client.Get(imprint.Link)
		if err != nil {
			imprint.Err = err
			return
		}
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
					imprint.Name = trim(lastLastValue)
					imprint.Address = trim(lastValue)
					imprint.Zip, imprint.City = ZipCityTrimmer(value)
				}
			}

			// check email
			if emailMatch == false {
				emailMatch, _ = regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", value)
				if emailMatch {
					imprint.Email = trim(value)
				}
			}

			lastLastValue = lastValue
			lastValue = value

		}
	}
}

func trim(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimLeft(s, " ")
	s = strings.TrimRight(s, " ")

	oldS := ""
	for oldS != s {
		oldS = s
		s = strings.ReplaceAll(oldS, "  ", " ")
	}

	return s
}

func concatLink(base string, ext string) string {
	u, _ := url.Parse(base)
	host := u.Hostname()

	sep := ""
	if strings.HasSuffix(host, "/") == false && strings.HasPrefix(ext, "/") == false {
		sep = "/"
	}
	if strings.HasSuffix(host, "/") && strings.HasPrefix(ext, "/") {
		host = strings.TrimRight(host, "/")
	}

	return u.Scheme + "://" + host + sep + ext
}

func chooseBestImprint(mainPage *MainPage) {
	var secondBest *Imprint
	for _, imprint := range mainPage.Imprints {
		if imprint.Zip != "" && imprint.Email != "" {
			mainPage.BestImprint = imprint
			return
		}
		if secondBest == nil && imprint.Zip != "" {
			secondBest = imprint
		}
	}

	for _, imprint := range mainPage.Contacts {
		if imprint.Zip != "" && imprint.Email != "" {
			mainPage.BestImprint = imprint
			return
		}
		if secondBest == nil && imprint.Zip != "" {
			secondBest = imprint
		}
	}

	mainPage.BestImprint = secondBest
}

func ZipCityTrimmer(zipLine string) (string, string) {
	r, err := regexp.Compile("[0-9][0-9][0-9][0-9][0-9] [A-Z][a-z]")
	if err != nil {
		panic(err)
	}
	indexes := r.FindStringSubmatchIndex(zipLine)
	if len(indexes) == 0 {
		panic("Missing indexes")
	}

	zipStartIndex := indexes[0]
	zip := zipLine[zipStartIndex : zipStartIndex+5]

	cityEndIndex := strings.Index(zipLine[zipStartIndex+6:], " ")
	if cityEndIndex == -1 {
		cityEndIndex = len(zipLine)
	} else {
		cityEndIndex = zipStartIndex + 6 + cityEndIndex
	}

	return zip, trim(zipLine[zipStartIndex+6 : cityEndIndex])
}
