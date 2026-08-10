package astisub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSSAOverrides(t *testing.T) {
	for _, v := range []struct {
		expected []ssaOverridePart
		name     string
		text     string
	}{
		{
			name:     "no override",
			text:     "Hello world",
			expected: []ssaOverridePart{{text: "Hello world"}},
		},
		{
			name: "leading override",
			text: `{\an8}Hello world`,
			expected: []ssaOverridePart{
				{tags: []ssaOverrideTag{{name: "an", args: "8"}}},
				{text: "Hello world"},
			},
		},
		{
			name: "in between override",
			text: `Hello {\i1}world`,
			expected: []ssaOverridePart{
				{text: "Hello "},
				{tags: []ssaOverrideTag{{name: "i", args: "1"}}},
				{text: "world"},
			},
		},
		{
			name: "names are not confused with one another",
			text: `{\an8\a6\b1\bord2\i1\iclip(0,0,1,1)}`,
			expected: []ssaOverridePart{{tags: []ssaOverrideTag{
				{name: "an", args: "8"},
				{name: "a", args: "6"},
				{name: "b", args: "1"},
				{name: "bord", args: "2"},
				{name: "i", args: "1"},
				{name: "iclip", args: "(0,0,1,1)"},
			}}},
		},
		{
			name: "color and alpha names keep their digit",
			text: `{\1c&H0000FF&\3a&HFF&}`,
			expected: []ssaOverridePart{{tags: []ssaOverrideTag{
				{name: "1c", args: "&H0000FF&"},
				{name: "3a", args: "&HFF&"},
			}}},
		},
		{
			name: "comments are dropped",
			text: `{\i1 this is a comment}`,
			expected: []ssaOverridePart{
				{tags: []ssaOverrideTag{{name: "i", args: "1 this is a comment"}}},
			},
		},
		{
			name:     "braces without tag are text",
			text:     "Hello {world}",
			expected: []ssaOverridePart{{text: "Hello {world}"}},
		},
		{
			name:     "unterminated brace is text",
			text:     `Hello {\i1 world`,
			expected: []ssaOverridePart{{text: `Hello {\i1 world`}},
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			assert.Equal(t, v.expected, splitSSAOverrides(v.text))
		})
	}
}

func TestApplySRTOverrideTags(t *testing.T) {
	var tags = func(block string) []ssaOverrideTag { return splitSSAOverrides(block)[0].tags }

	var sa = &StyleAttributes{}
	applySRTOverrideTags(sa, tags(`{\an8\b1\i1\u1\c&H0000FF&}`))
	assert.Equal(t, byte(8), sa.SRTPosition)
	assert.True(t, sa.SRTBold)
	assert.True(t, sa.SRTItalics)
	assert.True(t, sa.SRTUnderline)
	assert.Equal(t, &Color{Red: 255}, sa.SRTColor)

	// The tags that srt cannot represent are ignored, including the ones whose name
	// starts like a supported one
	applySRTOverrideTags(sa, tags(`{\bord2\iclip(0,0,1,1)\alpha&HFF&\clip(0,0,1,1)\an}`))
	assert.Equal(t, byte(8), sa.SRTPosition)
	assert.True(t, sa.SRTBold)
	assert.True(t, sa.SRTItalics)
	assert.Equal(t, &Color{Red: 255}, sa.SRTColor)

	applySRTOverrideTags(sa, tags(`{\b0}`))
	assert.False(t, sa.SRTBold)

	// \r reverts every tag
	applySRTOverrideTags(sa, tags(`{\r}`))
	assert.Equal(t, &StyleAttributes{}, sa)

	// The legacy alignment uses another layout
	applySRTOverrideTags(sa, tags(`{\a6}`))
	assert.Equal(t, byte(8), sa.SRTPosition)
	applySRTOverrideTags(sa, tags(`{\a10}`))
	assert.Equal(t, byte(5), sa.SRTPosition)
	applySRTOverrideTags(sa, tags(`{\a2}`))
	assert.Equal(t, byte(2), sa.SRTPosition)
}
