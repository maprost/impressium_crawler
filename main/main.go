package main

import (
	"fmt"

	"github.com/maprost/imprint_crawler"
)

func main() {
	//fmt.Println(imprint_crawler.CrawlMainPage("https://www.platanus-schule.de/"))
	//fmt.Println(imprint_crawler.CrawlMainPage("https://db.de"))
	fmt.Println(imprint_crawler.CrawlMainPage("https://www.deutschepost.de"))
}
