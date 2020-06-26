package imprint_crawler

import (
	"testing"

	"github.com/maprost/should"
)

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
