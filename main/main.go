package main

import (
	"github.com/maprost/imprint_crawler"
)

const (
	testVersion = 0
	liveVersion = 3
)

const version = testVersion

func main() {
	var path string

	switch version {
	case testVersion:
		path = "main/testVersion.txt"
	case liveVersion:
		path = "main/alleAnbieterUrlEmail.txt"
	}

	links, err := imprint_crawler.GetLinks(path)
	if err != nil {
		panic(err)
	}

	cache := imprint_crawler.CrawlMainPages(links, version)

	err = cache.CSV()
	if err != nil {
		panic(err)
	}

	err = cache.ErrorCSV()
	if err != nil {
		panic(err)
	}

	err = cache.Save()
	if err != nil {
		panic(err)
	}
}
