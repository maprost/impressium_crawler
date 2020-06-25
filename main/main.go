package main

import (
	"fmt"

	"github.com/maprost/imprint_crawler"
)

func main() {
	//fmt.Println(imprint_crawler.CrawlMainPage("https://www.platanus-schule.de/"))
	//fmt.Println(imprint_crawler.CrawlMainPage("https://db.de"))
	//fmt.Println(imprint_crawler.CrawlMainPage("https://www.deutschepost.de"))

	links, err := imprint_crawler.GetLinks("main/Lernorte2.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(links))

	for _, l := range links {
		fmt.Println(l)
	}

	cache := imprint_crawler.CrawlMainPages(links)
	cache.CSV()
}
