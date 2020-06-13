package impressium_crawler

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

var possibleTags = []string{"Kontakt"}

func CrawlList(links []string) {
	for _, link := range links {
		Crawl(link)
	}
}

func Crawl(link string) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println("error happen:", err.Error())
		return
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	for {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
		case tokenType == html.StartTagToken:
			token := z.Token()

			if token.Data == "a" {

				fmt.Println("start: ", z.Token().Data, z.Token().Attr)
			}
		}
	}
}

func imprintLink() {

}
