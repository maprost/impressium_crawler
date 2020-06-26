package imprint_crawler

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	jsonPath     = "main/cache"
	csvPath      = "main/cache"
	errorCsvPath = "main/error_cache"
)

type Cache struct {
	MainPages map[string]MainPage
	Version   int
}

func NewCache(version int) *Cache {
	c := &Cache{
		MainPages: make(map[string]MainPage),
		Version:   version,
	}

	c.Load()
	return c
}

func (c Cache) String() string {
	s := ""

	for _, p := range c.MainPages {
		s += p.String()
	}

	return s
}

func (c *Cache) Load() error {
	bytes, err := ioutil.ReadFile(jsonPath + c.versionToString() + ".json")
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

	return ioutil.WriteFile(jsonPath+c.versionToString()+".json", bytes, os.ModePerm)
}

func (c *Cache) CSV() error {
	s := CSVHeader()

	for _, p := range c.MainPages {
		if p.Err == nil {
			s += p.CSV()
		}
	}

	return ioutil.WriteFile(csvPath+c.versionToString()+".csv", []byte(s), os.ModePerm)
}

func (c *Cache) ErrorCSV() error {
	s := CSVHeader()

	for _, p := range c.MainPages {
		if p.Err != nil {
			s += p.CSV()
		}
	}

	return ioutil.WriteFile(errorCsvPath+c.versionToString()+".csv", []byte(s), os.ModePerm)
}
func (c *Cache) versionToString() string {
	return strconv.Itoa(c.Version)
}
