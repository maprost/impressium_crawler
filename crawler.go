package imprint_crawler

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"golang.org/x/net/html"
)

const (
	LinkRedirectDiffers = "LinkRedirectDiffers"
	goRoutines          = 50
)

var client = http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	},
}

type MainPage struct {
	Given    string
	Redirect string
	BaseUrl  string
	Title    TitleCheck
	Err      error
	Flag     string

	Address     AddressCheck
	Email       EMailCheck
	BestImprint *Imprint

	Imprints map[string]*Imprint
	Contacts map[string]*Imprint
}

func (p MainPage) String() string {
	s := fmt.Sprintln("Given: ", p.Given)
	s += fmt.Sprintln("Flag: ", p.Flag)
	s += fmt.Sprintln("Title: ", p.Title)
	s += fmt.Sprintln("Error: ", p.Err)
	s += fmt.Sprintln("Street: ", p.Address.Street())
	s += fmt.Sprintln("Zip: ", p.Address.Zip())
	s += fmt.Sprintln("City: ", p.Address.City())
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
	return fmt.Sprintln("Given,Error,Redirect,BaseUrl,Title,Street,Zip,City,Lat,Long,Email", ImprintCSVHeader())
}

func (p MainPage) CSV() string {
	errMsg := ""
	if p.Err != nil {
		errMsg = trim(p.Err.Error())
	}
	return fmt.Sprint(trimLinks(p.Given), ",", errMsg, ",", trimLinks(p.Redirect), ",", trimLinks(p.BaseUrl), ",", p.Title, ",", p.Address.Street(), ",", p.Address.Zip(), ",", p.Address.City(), ",", p.Address.Latitude(), ",", p.Address.Longitude(), ",", p.Email, p.BestImprint.CSV(), "\n")
}

type Imprint struct {
	Tag  string
	Link string
	Err  error

	Name    string
	Address AddressCheck
	Email   EMailCheck
}

func (i Imprint) String() string {
	s := fmt.Sprintln("\tTag/Link: ", i.Tag, i.Link)

	s += fmt.Sprintln("\tError: ", i.Err)

	s += fmt.Sprintln("\tName: ", i.Name)
	s += fmt.Sprintln("\tStreet: ", i.Address.Street())
	s += fmt.Sprintln("\tZip: ", i.Address.Zip())
	s += fmt.Sprintln("\tCity: ", i.Address.City())
	s += fmt.Sprintln("\tEmail: ", i.Email)

	return s
}

func ImprintCSVHeader() string {
	return fmt.Sprint(",Imprint-Url,Imprint-Street,Imprint-Zip,Imprint-City,Imprint-Lat,Imprint-Long,Imprint-Email")
}

func (i Imprint) CSV() string {
	return fmt.Sprint(",", trimLinks(i.Link), ",", i.Address.Street(), ",", i.Address.Zip(), ",", i.Address.City(), ",", i.Address.Latitude(), ",", i.Address.Longitude(), ",", i.Email)
}

func CrawlMainPages(links []string, version int) *Cache {
	cache := NewCache(version)
	cacheMutex := sync.Mutex{}
	counter := 0
	size := len(links)
	wg := sync.WaitGroup{}
	wg.Add(size)

	pool := tunny.NewFunc(goRoutines, func(payload interface{}) interface{} {
		var result []byte
		link := payload.(string)

		page := CrawlMainPage(link)

		cacheMutex.Lock()
		counter++
		fmt.Printf("\rProgress... %d/%d complete ", counter, size)
		cache.MainPages[link] = page
		wg.Done()
		cacheMutex.Unlock()

		return result
	})
	defer pool.Close()

	start := time.Now()
	for _, link := range links {
		go pool.Process(link)
	}

	wg.Wait()
	fmt.Println()
	fmt.Println("Process time:", time.Since(start))

	return cache
}

func CrawlMainPage(given string) MainPage {
	mainPage := MainPage{
		Given:       given,
		BestImprint: &Imprint{},
		Imprints:    make(map[string]*Imprint),
		Contacts:    make(map[string]*Imprint),
	}

	link := getLinkFromGiven(given)

	resp, err := client.Get(link)
	if err != nil {
		mainPage.Err = err
		return mainPage
	}

	mainPage.Redirect = resp.Request.URL.String()
	mainPage.BaseUrl = fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host)

	if mainPage.Redirect != link {
		r := strings.Replace(mainPage.Redirect, "http://", "", 1)
		r = strings.Replace(r, "http://", "", 1)

		l := strings.Replace(link, "http://", "", 1)
		l = strings.Replace(l, "http://", "", 1)

		if r != l {
			mainPage.Flag = LinkRedirectDiffers
		}
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

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
					blob := z.Token()
					_ = blob
					continue
				}
				tokenValue := z.Token()

				addImprint(mainPage, token, tokenValue)
			}

			mainPage.Title.check(token, z)

		case tokenType == html.TextToken:
			value := z.Token().String()

			mainPage.Address.check(value)
			mainPage.Email.check(value)
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

	for {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
			return

		case tokenType == html.TextToken:
			value := z.Token().String()

			imprint.Address.check(value)
			imprint.Email.check(value)
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
		if imprint.Address.Zip() != "" && imprint.Email.String() != "" {
			mainPage.BestImprint = imprint
			return
		}
		if secondBest == nil && imprint.Address.Zip() != "" {
			secondBest = imprint
		}
	}

	for _, imprint := range mainPage.Contacts {
		if imprint.Address.Zip() != "" && imprint.Email.String() != "" {
			mainPage.BestImprint = imprint
			return
		}
		if secondBest == nil && imprint.Address.Zip() != "" {
			secondBest = imprint
		}
	}

	if secondBest == nil {
		secondBest = &Imprint{}
	}
	mainPage.BestImprint = secondBest
}

func trimLinks(link string) string {
	return strings.Replace(link, ",", "%2C", -1)
}

func getLinkFromGiven(given string) string {
	if atIndex := strings.Index(given, "@"); atIndex != -1 {
		return "http://www." + given[atIndex+1:]
	}
	return given
}
