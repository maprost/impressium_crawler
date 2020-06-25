package imprint_crawler

import (
	"testing"

	"github.com/maprost/should"
)

func TestZipTrimmer(t *testing.T) {
	t.Run("zip + city", func(t *testing.T) {
		zip, city := ZipCityTrimmer("12345 Freiberg")
		should.BeEqual(t, zip, "12345")
		should.BeEqual(t, city, "Freiberg")
	})

	t.Run("random string + zip + city", func(t *testing.T) {
		zip, city := ZipCityTrimmer("kf kfgbdjk  12345 Freiberg")
		should.BeEqual(t, zip, "12345")
		should.BeEqual(t, city, "Freiberg")
	})

	t.Run("zip + city + random string", func(t *testing.T) {
		zip, city := ZipCityTrimmer("12345 Freiberg kf kfgbdj")
		should.BeEqual(t, zip, "12345")
		should.BeEqual(t, city, "Freiberg")
	})
}

func TestTrim(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		should.BeEqual(t, trim("blob?"), "blob?")
	})

	t.Run("more spaces in the middle", func(t *testing.T) {
		should.BeEqual(t, trim("blob     ?"), "blob ?")
	})
}

func TestGetLinkFromGiven(t *testing.T) {
	t.Run("url", func(t *testing.T) {
		should.BeEqual(t, getLinkFromGiven("blob.de"), "blob.de")
	})

	t.Run("email", func(t *testing.T) {
		should.BeEqual(t, getLinkFromGiven("jo@blob.de"), "blob.de")
	})
}
