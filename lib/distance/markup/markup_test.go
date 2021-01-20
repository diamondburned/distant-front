package markup

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestMarkup(t *testing.T) {
	type test struct {
		Input string
		autogold.Value
	}

	// go test -update
	var tests = []test{{
		"[00FF00]Tip:[-] Use [00FFFF]/search[-] to search for levels before voting",
		autogold.Want("partial coloring", `<span style="color:#00FF00">Tip:</span> Use <span style="color:#00FFFF">/search</span> to search for levels before voting`),
	}, {
		"[FFE999]XERU has joined the server!",
		autogold.Want("missing close tag", `<span style="color:#FFE999">XERU has joined the server!</span>`),
	}, {
		"[c][FFE999]Rynero reset[-][/c]",
		autogold.Want("unknown c tag", `<span style="color:#FFE999">Rynero reset</span>`),
	}, {
		"[url=https://google.com]best website[/url]",
		autogold.Want("hyperlink", `<a href="https://google.com">best website</a>`),
	}}

	for _, test := range tests {
		t.Run(test.Name(), func(t *testing.T) {
			test.Value.Equal(t, ToHTML(test.Input))
		})
	}
}
