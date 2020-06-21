package imprint_crawler

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const path = "cache.json"

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
	bytes, err := ioutil.ReadFile(path)
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

	return ioutil.WriteFile(path, bytes, os.ModePerm)
}
