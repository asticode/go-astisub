package astisub

import (
	"strconv"
	"testing"
	"time"

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

	t.Run("When inline timestamps are included", func(t *testing.T) {
		testData := `<00:01:01.000>With inline <00:01:02.000>timestamps`

		s := parseTextWebVTT(testData)
		assert.Equal(t, 2, len(s.Items))
		assert.Equal(t, "With inline", s.Items[0].Text)
		assert.Equal(t, time.Minute+time.Second, s.Items[0].StartAt)
		assert.Equal(t, "timestamps", s.Items[1].Text)
		assert.Equal(t, time.Minute+2*time.Second, s.Items[1].StartAt)
	})

	t.Run("When inline timestamps together", func(t *testing.T) {
		testData := `<00:01:01.000><00:01:02.000>With timestamp tags together`

		s := parseTextWebVTT(testData)
		assert.Equal(t, 1, len(s.Items))
		assert.Equal(t, "With timestamp tags together", s.Items[0].Text)
		assert.Equal(t, time.Minute+2*time.Second, s.Items[0].StartAt)
	})

	t.Run("When inline timestamps is at end", func(t *testing.T) {
		testData := `With end timestamp<00:01:02.000>`

		s := parseTextWebVTT(testData)
		assert.Equal(t, 1, len(s.Items))
		assert.Equal(t, "With end timestamp", s.Items[0].Text)
		assert.Equal(t, time.Duration(0), s.Items[0].StartAt)
	})
}

func TestTimestampMap(t *testing.T) {
	for i, c := range []struct {
		line           string
		expectedOffset time.Duration
		expectError    bool
	}{
		{
			line:           "X-TIMESTAMP-MAP=MPEGTS:180000, LOCAL:00:00:00.000",
			expectedOffset: 2 * time.Second,
		},
		{
			line:           "X-TIMESTAMP-MAP=MPEGTS:180000, LOCAL:00:00:00.500",
			expectedOffset: 1500 * time.Millisecond,
		},
		{
			line:           "X-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:135000",
			expectedOffset: 1500 * time.Millisecond,
		},
		{
			line:           "X-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:324090000",
			expectedOffset: time.Hour + time.Second,
		},
		{
			line:        "X-TIMESTAMP-MAP=MPEGTS:foo, LOCAL:00:00:00.000",
			expectError: true,
		},
		{
			line:        "X-TIMESTAMP-MAP=MPEGTS:180000,LOCAL:bar",
			expectError: true,
		},
		{
			line:        "X-TIMESTAMP-MAP=MPEGTS:180000,LOCAL",
			expectError: true,
		},
		{
			line:        "X-TIMESTAMP-MAP=MPEGTS,LOCAL:00:00:00.000",
			expectError: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			timestampMap, err := parseTimestampMapWebVTT(c.line)
			offset := timestampMap.Offset()
			assert.Equal(t, c.expectedOffset, offset)
			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCueVoiceSpanRegex(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{
			give: `<v 中文> this is the content</v>`,
			want: `中文`,
		},
		{
			give: `<v 中文> this is the content`,
			want: `中文`,
		},
		{
			give: `<v.abc 中文> this is the content</v>`,
			want: `中文`,
		},
		{
			give: `<v.jp 言語の> this is the content`,
			want: `言語の`,
		},
		{
			give: `<v.ko 언어> this is the content`,
			want: `언어`,
		},
		{
			give: `<v foo bar> this is the content`,
			want: `foo bar`,
		},
		{
			give: `<v هذا عربي> this is the content`,
			want: `هذا عربي`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			results := webVTTRegexpTag.FindStringSubmatch(tt.give)
			assert.True(t, len(results) == 5)
			assert.Equal(t, tt.want, results[4])
		})
	}
}
