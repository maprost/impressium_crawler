package imprint_crawler

import (
	"testing"

	"github.com/maprost/should"
)

func TestZipTrimmer(t *testing.T) {
	should.BeEqual(t, ZipTrimmer("kf kfgbdjk  12345 Freiberg"), "12345")
}
