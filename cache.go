package imprint_crawler

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const (
	jsonPath = "cache.json"
	csvPath  = "cache.csv"
)

type Cache struct {
	MainPages map[string]MainPage
}

func NewCache() *Cache {
	c := &Cache{
		MainPages: make(map[string]MainPage),
	}

	c.Load()
	return c
}

func (c *Cache) Load() error {
	bytes, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

func (c *Cache) Save() error {
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(jsonPath, bytes, os.ModePerm)
}

func (c *Cache) CSV() error {
	s := CSVHeader()

	for _, p := range c.MainPages {
		s += p.CSV()
	}

	return ioutil.WriteFile(csvPath, []byte(s), os.ModePerm)
}
