package astisub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTextWebVTT(t *testing.T) {

	t.Run("When both voice tags are available", func(t *testing.T) {
		testData := `<v Bob>Correct tag</v>`

		s := parseTextWebVTT(testData)
		assert.Equal(t, "Bob", s.VoiceName)
		assert.Equal(t, 1, len(s.Items))
		assert.Equal(t, "Correct tag", s.Items[0].Text)
	})

	t.Run("When there is no end tag", func(t *testing.T) {
		testData := `<v Bob> Text without end tag`

		s := parseTextWebVTT(testData)
		assert.Equal(t, "Bob", s.VoiceName)
		assert.Equal(t, 1, len(s.Items))
		assert.Equal(t, "Text without end tag", s.Items[0].Text)
	})

	t.Run("When the end tag is correct", func(t *testing.T) {
		testData := `<v Bob>Incorrect end tag</vi>`

		s := parseTextWebVTT(testData)
		assert.Equal(t, "Bob", s.VoiceName)
		assert.Equal(t, 1, len(s.Items))
		assert.Equal(t, "Incorrect end tag", s.Items[0].Text)
	})
}
