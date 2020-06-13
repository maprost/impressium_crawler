package impressium_crawler

func CrawlList(links []string) {
	for _, link := range links {
		Crawl(link)
	}
}

func Crawl(link string) {

}
