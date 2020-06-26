package imprint_crawler

import (
	"testing"

	"github.com/maprost/should"
)

func TestAddressCheck(t *testing.T) {
	t.Run("zip + city", func(t *testing.T) {
		check := AddressCheck{}
		check.check("12345 Freiberg")
		should.BeEqual(t, check.Zip(), "12345")
		should.BeEqual(t, check.City(), "Freiberg")
	})

	t.Run("random string + zip + city", func(t *testing.T) {
		check := AddressCheck{}
		check.check("kf kfgbdjk  12345 Freiberg")
		should.BeEqual(t, check.Zip(), "12345")
		should.BeEqual(t, check.City(), "Freiberg")
	})

	t.Run("zip + city + random string", func(t *testing.T) {
		check := AddressCheck{}
		check.check("12345 Freiberg kf kfgbdj")
		should.BeEqual(t, check.Zip(), "12345")
		should.BeEqual(t, check.City(), "Freiberg")
	})

	t.Run("take first one", func(t *testing.T) {
		check := AddressCheck{}

		check.check("12345 Freiberg")
		should.BeEqual(t, check.Zip(), "12345")
		should.BeEqual(t, check.City(), "Freiberg")

		check.check("55555 Gortig")
		should.BeEqual(t, check.Zip(), "12345")
		should.BeEqual(t, check.City(), "Freiberg")
	})

	t.Run("check street without number", func(t *testing.T) {
		check := AddressCheck{}

		check.check("Tiroler Straße")
		check.check("13187 Berlin")
		should.BeEqual(t, check.Street(), "")
	})

	t.Run("check longitut and latitute", func(t *testing.T) {
		check := AddressCheck{}

		check.check("Tiroler Straße 60")
		check.check("13187 Berlin")
		should.BeEqual(t, check.Street(), "Tiroler Straße 60")
		should.BeEqual(t, check.Longitude(), "13.417100")
		should.BeEqual(t, check.Latitude(), "52.573100")
	})
}

func TestEMailCheck(t *testing.T) {
	t.Run("simple email", func(t *testing.T) {
		emailCheck := EMailCheck{}

		emailCheck.check("foo@world.de")
		should.BeEqual(t, emailCheck.String(), "foo@world.de")
	})

	t.Run("take first one", func(t *testing.T) {
		emailCheck := EMailCheck{}

		emailCheck.check("foo@world.de")
		should.BeEqual(t, emailCheck.String(), "foo@world.de")

		emailCheck.check("bar@world.de")
		should.BeEqual(t, emailCheck.String(), "foo@world.de")
	})

	t.Run("no dot", func(t *testing.T) {
		emailCheck := EMailCheck{}

		emailCheck.check("foo@world")
		should.BeEqual(t, emailCheck.String(), "foo@world")
	})

	t.Run("(at)", func(t *testing.T) {
		emailCheck := EMailCheck{}

		emailCheck.check("foo(at)world.de")
		should.BeEqual(t, emailCheck.String(), "foo@world.de")
	})

	t.Run("real example", func(t *testing.T) {
		emailCheck := EMailCheck{}

		emailCheck.check("kontakt@gedenkstaette-hoheneck.com")
		should.BeEqual(t, emailCheck.String(), "kontakt@gedenkstaette-hoheneck.com")
	})
}
