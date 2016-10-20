package s2prot

import (
	"testing"

	"github.com/icza/s2prot/build"
)

func TestGetProtocol(t *testing.T) {
	for baseBuild, _ := range build.Builds {
		p := GetProtocol(baseBuild)
		if p == nil {
			t.Errorf("Parsing protocol %d failed!", baseBuild)
		}
	}
}

func BenchmarkParseProtocol(b *testing.B) {
	//baseBuild := 15405
	baseBuild := 47185

	for i := 0; i < b.N; i++ {
		parseProtocol(build.Builds[baseBuild], baseBuild)
	}
}
