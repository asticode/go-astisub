package astisub_test

import (
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestLine_Text(t *testing.T) {
	var l = astisub.Line{{Text: "1"}, {Text: "2"}, {Text: "3"}}
	assert.Equal(t, "1 2 3", l.String())
}

func assertSubtitleItems(t *testing.T, i *astisub.Subtitles) {
	// No format
	assert.Len(t, i.Items, 6)
	assert.Equal(t, time.Minute+39*time.Second, i.Items[0].StartAt)
	assert.Equal(t, time.Minute+41*time.Second+40*time.Millisecond, i.Items[0].EndAt)
	assert.Equal(t, "(deep rumbling)", i.Items[0].Lines[0].String())
	assert.Equal(t, 2*time.Minute+4*time.Second+80*time.Millisecond, i.Items[1].StartAt)
	assert.Equal(t, 2*time.Minute+7*time.Second+120*time.Millisecond, i.Items[1].EndAt)
	assert.Equal(t, "MAN:", i.Items[1].Lines[0].String())
	assert.Equal(t, "How did we end up here?", i.Items[1].Lines[1].String())
	assert.Equal(t, 2*time.Minute+12*time.Second+160*time.Millisecond, i.Items[2].StartAt)
	assert.Equal(t, 2*time.Minute+15*time.Second+200*time.Millisecond, i.Items[2].EndAt)
	assert.Equal(t, "This place is horrible.", i.Items[2].Lines[0].String())
	assert.Equal(t, 2*time.Minute+20*time.Second+240*time.Millisecond, i.Items[3].StartAt)
	assert.Equal(t, 2*time.Minute+22*time.Second+280*time.Millisecond, i.Items[3].EndAt)
	assert.Equal(t, "Smells like balls.", i.Items[3].Lines[0].String())
	assert.Equal(t, 2*time.Minute+28*time.Second+320*time.Millisecond, i.Items[4].StartAt)
	assert.Equal(t, 2*time.Minute+31*time.Second+360*time.Millisecond, i.Items[4].EndAt)
	assert.Equal(t, "We don't belong", i.Items[4].Lines[0].String())
	assert.Equal(t, "in this shithole.", i.Items[4].Lines[1].String())
	assert.Equal(t, 2*time.Minute+31*time.Second+400*time.Millisecond, i.Items[5].StartAt)
	assert.Equal(t, 2*time.Minute+33*time.Second+440*time.Millisecond, i.Items[5].EndAt)
	assert.Equal(t, "(computer playing", i.Items[5].Lines[0].String())
	assert.Equal(t, "electronic melody)", i.Items[5].Lines[1].String())
}

func mockSubtitles() *astisub.Subtitles {
	return &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 3 * time.Second, StartAt: time.Second, Lines: []astisub.Line{{{Text: "subtitle-1"}}}}, {EndAt: 7 * time.Second, StartAt: 3 * time.Second, Lines: []astisub.Line{{{Text: "subtitle-2"}}}}}}
}

func TestSubtitles_Add(t *testing.T) {
	var s = mockSubtitles()
	s.Add(time.Second)
	assert.Len(t, s.Items, 2)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
}

func TestSubtitles_Duration(t *testing.T) {
	assert.Equal(t, time.Duration(0), astisub.Subtitles{}.Duration())
	assert.Equal(t, 7*time.Second, mockSubtitles().Duration())
}

func TestSubtitles_IsEmpty(t *testing.T) {
	assert.True(t, astisub.Subtitles{}.IsEmpty())
	assert.False(t, mockSubtitles().IsEmpty())
}

func TestSubtitles_ForceDuration(t *testing.T) {
	var s = mockSubtitles()
	s.ForceDuration(10 * time.Second)
	assert.Len(t, s.Items, 3)
	assert.Equal(t, 10*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 10*time.Second, s.Items[2].StartAt)
	assert.Equal(t, []astisub.Line{{{Text: "..."}}}, s.Items[2].Lines)
	s.Items[2].StartAt = 7 * time.Second
	s.Items[2].EndAt = 12 * time.Second
	s.ForceDuration(10 * time.Second)
	assert.Len(t, s.Items, 3)
	assert.Equal(t, 10*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 7*time.Second, s.Items[2].StartAt)
}

func TestSubtitles_Fragment(t *testing.T) {
	// Init
	var s = mockSubtitles()

	// Fragment
	s.Fragment(2 * time.Second)
	assert.Len(t, s.Items, 5)
	assert.Equal(t, time.Second, s.Items[0].StartAt)
	assert.Equal(t, 2*time.Second, s.Items[0].EndAt)
	assert.Equal(t, []astisub.Line{{{Text: "subtitle-1"}}}, s.Items[0].Lines)
	assert.Equal(t, 2*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[1].EndAt)
	assert.Equal(t, []astisub.Line{{{Text: "subtitle-1"}}}, s.Items[1].Lines)
	assert.Equal(t, 3*time.Second, s.Items[2].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[2].EndAt)
	assert.Equal(t, []astisub.Line{{{Text: "subtitle-2"}}}, s.Items[2].Lines)
	assert.Equal(t, 4*time.Second, s.Items[3].StartAt)
	assert.Equal(t, 6*time.Second, s.Items[3].EndAt)
	assert.Equal(t, []astisub.Line{{{Text: "subtitle-2"}}}, s.Items[3].Lines)
	assert.Equal(t, 6*time.Second, s.Items[4].StartAt)
	assert.Equal(t, 7*time.Second, s.Items[4].EndAt)
	assert.Equal(t, []astisub.Line{{{Text: "subtitle-2"}}}, s.Items[4].Lines)

	// Unfragment
	s.Unfragment()
	assert.Len(t, s.Items, 2)
	assert.Equal(t, "subtitle-1", s.Items[0].String())
	assert.Equal(t, time.Second, s.Items[0].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[0].EndAt)
	assert.Equal(t, "subtitle-2", s.Items[1].String())
	assert.Equal(t, 3*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 7*time.Second, s.Items[1].EndAt)
}

func TestSubtitles_Merge(t *testing.T) {
	var s1 = &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 3 * time.Second, StartAt: time.Second}, {EndAt: 8 * time.Second, StartAt: 5 * time.Second}, {EndAt: 12 * time.Second, StartAt: 10 * time.Second}}}
	var s2 = &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, {EndAt: 7 * time.Second, StartAt: 6 * time.Second}, {EndAt: 11 * time.Second, StartAt: 9 * time.Second}, {EndAt: 14 * time.Second, StartAt: 13 * time.Second}}}
	s1.Merge(s2)
	assert.Len(t, s1.Items, 7)
	assert.Equal(t, &astisub.Item{EndAt: 3 * time.Second, StartAt: time.Second}, s1.Items[0])
	assert.Equal(t, &astisub.Item{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, s1.Items[1])
	assert.Equal(t, &astisub.Item{EndAt: 8 * time.Second, StartAt: 5 * time.Second}, s1.Items[2])
	assert.Equal(t, &astisub.Item{EndAt: 7 * time.Second, StartAt: 6 * time.Second}, s1.Items[3])
	assert.Equal(t, &astisub.Item{EndAt: 11 * time.Second, StartAt: 9 * time.Second}, s1.Items[4])
	assert.Equal(t, &astisub.Item{EndAt: 12 * time.Second, StartAt: 10 * time.Second}, s1.Items[5])
	assert.Equal(t, &astisub.Item{EndAt: 14 * time.Second, StartAt: 13 * time.Second}, s1.Items[6])
}

func TestSubtitles_Order(t *testing.T) {
	var s = &astisub.Subtitles{Items: []*astisub.Item{{StartAt: 4 * time.Second, EndAt: 5 * time.Second}, {StartAt: 2 * time.Second, EndAt: 3 * time.Second}, {StartAt: 3 * time.Second, EndAt: 4 * time.Second}, {StartAt: time.Second, EndAt: 2 * time.Second}}}
	s.Order()
	assert.Equal(t, time.Second, s.Items[0].StartAt)
	assert.Equal(t, 2*time.Second, s.Items[0].EndAt)
	assert.Equal(t, 2*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[1].EndAt)
	assert.Equal(t, 3*time.Second, s.Items[2].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 4*time.Second, s.Items[3].StartAt)
	assert.Equal(t, 5*time.Second, s.Items[3].EndAt)
}
