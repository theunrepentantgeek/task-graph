package indentwriter

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestWordWrap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		text    string
		width   int
		results []string
	}{
		{
			"this is a simple line of text",
			15,
			[]string{"this is a ", "simple line of ", "text"},
		},
		{
			"this is a simple line of text",
			16,
			[]string{"this is a simple ", "line of text"},
		},
		{
			"this is a simple line of text",
			20,
			[]string{"this is a simple ", "line of text"},
		},
		{
			"this is a simple line of text",
			21,
			[]string{"this is a simple line ", "of text"},
		},
		{
			"",
			0,
			[]string{},
		},
		{
			"this is a sample text",
			0,
			[]string{"this ", "is ", "a ", "sample ", "text"},
		},
	}

	for _, c := range cases {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)
			lines := WordWrap(c.text, c.width)
			g.Expect(lines).To(Equal(c.results))
		})
	}
}
