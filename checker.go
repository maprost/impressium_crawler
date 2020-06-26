package imprint_crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// ######################## Title #############################

type TitleCheck struct {
	match bool
	title string
}

func (c *TitleCheck) check(token html.Token, z *html.Tokenizer) {
	if c.match == false && strings.ToLower(token.Data) == "title" {
		// need the '<title>..</title>' token value
		if z.Next() != html.TextToken {
			return
		}

		tokenValue := z.Token()
		c.title = trim(tokenValue.String())
		c.match = true
	}
}

func (c TitleCheck) String() string {
	return c.title
}

// ######################## E-Mail #############################

var eMailCheckRegEx1, _ = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
var eMailCheckRegEx2, _ = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+\\(at\\)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type EMailCheck struct {
	match bool
	email string
}

func (c *EMailCheck) check(value string) {
	if c.match == false {
		c.match = eMailCheckRegEx1.MatchString(value)
		if c.match {
			c.email = trim(value)
		}
	}

	if c.match == false {
		c.match = eMailCheckRegEx2.MatchString(value)
		if c.match {
			value := strings.Replace(value, "(at)", "@", 1)
			c.email = trim(value)
		}
	}
}

func (c EMailCheck) String() string {
	return c.email
}

// ######################## Address #############################

var addressCheckRegEx, _ = regexp.Compile("[0-9][0-9][0-9][0-9][0-9] [A-Z][a-z]")

type AddressCheck struct {
	match     bool
	lastValue string
	street    string
	zip       string
	city      string
	longitude string
	latitude  string
}

func (c *AddressCheck) check(value string) {
	if c.match == false {
		c.match = addressCheckRegEx.MatchString(value)
		if c.match {
			c.addStreet()
			c.addZipCity(value)
			c.addLonLat()
		}
	}

	c.lastValue = value
}

func (c *AddressCheck) addStreet() {
	street := trim(c.lastValue)

	// a street should contain a number (we ignore the few one without a number to remove the wrong entries)
	if strings.ContainsAny(street, "0123456789") {
		c.street = street
	}
}

func (c *AddressCheck) addZipCity(zipLine string) {
	indexes := addressCheckRegEx.FindStringSubmatchIndex(zipLine)
	if len(indexes) == 0 {
		panic("Missing indexes")
	}

	zipStartIndex := indexes[0]
	zip := zipLine[zipStartIndex : zipStartIndex+5]

	cityEndIndex := strings.Index(zipLine[zipStartIndex+6:], " ")
	if cityEndIndex == -1 {
		cityEndIndex = len(zipLine)
	} else {
		cityEndIndex = zipStartIndex + 6 + cityEndIndex
	}

	c.zip = zip
	c.city = trim(zipLine[zipStartIndex+6 : cityEndIndex])
}

func (c AddressCheck) Street() string {
	return c.street
}

func (c AddressCheck) Zip() string {
	return c.zip
}

func (c AddressCheck) City() string {
	return c.city
}

func (c AddressCheck) Longitude() string {
	return c.longitude
}
func (c AddressCheck) Latitude() string {
	return c.latitude
}

type LonLatResult struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64
				Lng float64
			}
		}
	}
	Status string
}

func (c *AddressCheck) addLonLat() {
	// only calculate the longitude and latitude if we have all the infos
	if c.street == "" || c.zip == "" {
		return
	}

	street := strings.Replace(c.street, " ", "+", -1)
	link := fmt.Sprintf("http://www.datasciencetoolkit.org/maps/api/geocode/json?sensor=false&address=%s,%s,Germany", street, c.zip)

	resp, err := client.Get(link)
	if err != nil {
		fmt.Println("http error:", err)
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("body error:", err)
		return
	}

	var result LonLatResult
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		fmt.Println("json error:", err)
		return
	}

	if len(result.Results) > 0 {
		location := result.Results[0].Geometry.Location

		// api returns sometimes a wrong result, the long and lat are based in Germany so has to be positive
		if location.Lat > 0 && location.Lng > 0 {
			c.longitude = fmt.Sprintf("%f", location.Lng)
			c.latitude = fmt.Sprintf("%f", location.Lat)
		}
	}
}
