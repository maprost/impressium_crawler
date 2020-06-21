package imprint_crawler

import (
	"io/ioutil"
	"strings"
)

func GetLinks(path string) ([]string, error) {
	links := make([]string, 0, 10)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return links, err
	}

	s := strings.ReplaceAll(string(bytes), "\r", "")
	links = strings.Split(s, "\n")

	return links, nil
}
