package s2prot

import (
	"github.com/icza/s2prot/build"
	"testing"
)

func TestGetProtocol(t *testing.T) {
	cases := []struct {
		baseBuild int
	}{
		{15405},
		{34835},
	}

	for _, c := range cases {
		p := GetProtocol(c.baseBuild)
		if p == nil {
			t.Errorf("Parsing protocol %d failed!", c.baseBuild)
		}
	}
}

func BenchmarkParseProtocol(b *testing.B) {
	//baseBuild := 15405
	baseBuild := 34835

	for i := 0; i < b.N; i++ {
		parseProtocol(build.Builds[baseBuild], baseBuild)
	}
}
