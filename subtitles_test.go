package astisub_test

import (
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLine_Text(t *testing.T) {
	var l = astisub.Line{Items: []astisub.LineItem{{Text: "1"}, {Text: "2"}, {Text: "3"}}}
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
	return &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 3 * time.Second, StartAt: time.Second, Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-1"}}}}}, {EndAt: 7 * time.Second, StartAt: 3 * time.Second, Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-2"}}}}}}}
}

func TestSubtitles_Add(t *testing.T) {
	var s = mockSubtitles()
	s.Add(time.Second)
	assert.Len(t, s.Items, 2)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
	s.Add(-3 * time.Second)
	assert.Len(t, s.Items, 2)
	assert.Equal(t, time.Duration(0), s.Items[0].StartAt)
	assert.Equal(t, time.Second, s.Items[0].EndAt)
	s.Add(-2 * time.Second)
	assert.Len(t, s.Items, 1)
	assert.Equal(t, "subtitle-2", s.Items[0].Lines[0].Items[0].Text)
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
	s.ForceDuration(10*time.Second, false)
	assert.Len(t, s.Items, 2)
	s = mockSubtitles()
	s.ForceDuration(10*time.Second, true)
	assert.Len(t, s.Items, 3)
	assert.Equal(t, 10*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 10*time.Second-time.Millisecond, s.Items[2].StartAt)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "..."}}}}, s.Items[2].Lines)
	s.Items[2].StartAt = 7 * time.Second
	s.Items[2].EndAt = 12 * time.Second
	s.ForceDuration(10*time.Second, true)
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
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-1"}}}}, s.Items[0].Lines)
	assert.Equal(t, 2*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[1].EndAt)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-1"}}}}, s.Items[1].Lines)
	assert.Equal(t, 3*time.Second, s.Items[2].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[2].EndAt)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-2"}}}}, s.Items[2].Lines)
	assert.Equal(t, 4*time.Second, s.Items[3].StartAt)
	assert.Equal(t, 6*time.Second, s.Items[3].EndAt)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-2"}}}}, s.Items[3].Lines)
	assert.Equal(t, 6*time.Second, s.Items[4].StartAt)
	assert.Equal(t, 7*time.Second, s.Items[4].EndAt)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-2"}}}}, s.Items[4].Lines)

	// Unfragment
	s.Items = append(s.Items[:4], append([]*astisub.Item{{EndAt: 5 * time.Second, Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: "subtitle-3"}}}}, StartAt: 4 * time.Second}}, s.Items[4:]...)...)
	s.Unfragment()
	assert.Len(t, s.Items, 3)
	assert.Equal(t, "subtitle-1", s.Items[0].String())
	assert.Equal(t, time.Second, s.Items[0].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[0].EndAt)
	assert.Equal(t, "subtitle-2", s.Items[1].String())
	assert.Equal(t, 3*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 7*time.Second, s.Items[1].EndAt)
	assert.Equal(t, "subtitle-3", s.Items[2].String())
	assert.Equal(t, 4*time.Second, s.Items[2].StartAt)
	assert.Equal(t, 5*time.Second, s.Items[2].EndAt)
}

func TestSubtitles_Unfragment(t *testing.T) {
	itemText := func(s string) []astisub.Line {
		return []astisub.Line{{Items: []astisub.LineItem{{Text: s}}}}
	}
	items := []*astisub.Item{{
		Lines:   itemText("subtitle-1"),
		StartAt: 1 * time.Second,
		EndAt:   2 * time.Second,
	}, {
		Lines:   itemText("subtitle-2"),
		StartAt: 2 * time.Second,
		EndAt:   5 * time.Second,
	}, {
		Lines:   itemText("subtitle-3"),
		StartAt: 3 * time.Second,
		EndAt:   4 * time.Second,
	}, {
		// gap and nested within first subtitle-2; should not override end time
		Lines:   itemText("subtitle-2"),
		StartAt: 3 * time.Second,
		EndAt:   4 * time.Second,
	}, {
		Lines: itemText("subtitle-3"),
		// gap and start time equals previous end time
		StartAt: 4 * time.Second,
		EndAt:   5 * time.Second,
	}, {
		// should not be combined
		Lines:   itemText("subtitle-3"),
		StartAt: 6 * time.Second,
		EndAt:   7 * time.Second,
	}, {
		// test correcting for out-of-orderness
		Lines:   itemText("subtitle-1"),
		StartAt: 0 * time.Second,
		EndAt:   3 * time.Second,
	}}

	s := &astisub.Subtitles{Items: items}

	s.Unfragment()

	expected := []astisub.Item{{
		Lines:   itemText("subtitle-1"),
		StartAt: 0 * time.Second,
		EndAt:   3 * time.Second,
	}, {
		Lines:   itemText("subtitle-2"),
		StartAt: 2 * time.Second,
		EndAt:   5 * time.Second,
	}, {
		Lines:   itemText("subtitle-3"),
		StartAt: 3 * time.Second,
		EndAt:   5 * time.Second,
	}, {
		Lines:   itemText("subtitle-3"),
		StartAt: 6 * time.Second,
		EndAt:   7 * time.Second,
	}}

	assert.Equal(t, len(expected), len(s.Items))
	for i := range expected {
		assert.Equal(t, expected[i], *s.Items[i])
	}
}

func TestSubtitles_Merge(t *testing.T) {
	var s1 = &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 3 * time.Second, StartAt: time.Second}, {EndAt: 8 * time.Second, StartAt: 5 * time.Second}, {EndAt: 12 * time.Second, StartAt: 10 * time.Second}}, Regions: map[string]*astisub.Region{"region_0": {ID: "region_0"}, "region_1": {ID: "region_1"}}, Styles: map[string]*astisub.Style{"style_0": {ID: "style_0"}, "style_1": {ID: "style_1"}}}
	var s2 = &astisub.Subtitles{Items: []*astisub.Item{{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, {EndAt: 7 * time.Second, StartAt: 6 * time.Second}, {EndAt: 11 * time.Second, StartAt: 9 * time.Second}, {EndAt: 14 * time.Second, StartAt: 13 * time.Second}}, Regions: map[string]*astisub.Region{"region_1": {ID: "region_1"}, "region_2": {ID: "region_2"}}, Styles: map[string]*astisub.Style{"style_1": {ID: "style_1"}, "style_2": {ID: "style_2"}}}
	s1.Merge(s2)
	assert.Len(t, s1.Items, 7)
	assert.Equal(t, &astisub.Item{EndAt: 3 * time.Second, StartAt: time.Second}, s1.Items[0])
	assert.Equal(t, &astisub.Item{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, s1.Items[1])
	assert.Equal(t, &astisub.Item{EndAt: 8 * time.Second, StartAt: 5 * time.Second}, s1.Items[2])
	assert.Equal(t, &astisub.Item{EndAt: 7 * time.Second, StartAt: 6 * time.Second}, s1.Items[3])
	assert.Equal(t, &astisub.Item{EndAt: 11 * time.Second, StartAt: 9 * time.Second}, s1.Items[4])
	assert.Equal(t, &astisub.Item{EndAt: 12 * time.Second, StartAt: 10 * time.Second}, s1.Items[5])
	assert.Equal(t, &astisub.Item{EndAt: 14 * time.Second, StartAt: 13 * time.Second}, s1.Items[6])
	assert.Equal(t, len(s1.Regions), 3)
	assert.Equal(t, len(s1.Styles), 3)
}

func TestSubtitles_Optimize(t *testing.T) {
	var s = &astisub.Subtitles{
		Items: []*astisub.Item{
			{Region: &astisub.Region{ID: "1"}},
			{Style: &astisub.Style{ID: "1"}},
			{Lines: []astisub.Line{{Items: []astisub.LineItem{{Style: &astisub.Style{ID: "2"}}}}}},
		},
		Regions: map[string]*astisub.Region{
			"1": {ID: "1", Style: &astisub.Style{ID: "3"}},
			"2": {ID: "2", Style: &astisub.Style{ID: "4"}},
		},
		Styles: map[string]*astisub.Style{
			"1": {ID: "1"},
			"2": {ID: "2"},
			"3": {ID: "3"},
			"4": {ID: "4"},
			"5": {ID: "5"},
		},
	}
	s.Optimize()
	assert.Len(t, s.Regions, 1)
	assert.Len(t, s.Styles, 3)
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

func TestSubtitles_RemoveStyling(t *testing.T) {
	s := &astisub.Subtitles{
		Items: []*astisub.Item{
			{
				Lines: []astisub.Line{{
					Items: []astisub.LineItem{{
						InlineStyle: &astisub.StyleAttributes{},
						Style:       &astisub.Style{},
					}},
				}},
				InlineStyle: &astisub.StyleAttributes{},
				Region:      &astisub.Region{},
				Style:       &astisub.Style{},
			},
		},
		Regions: map[string]*astisub.Region{"region": {}},
		Styles:  map[string]*astisub.Style{"style": {}},
	}
	s.RemoveStyling()
	assert.Equal(t, &astisub.Subtitles{
		Items: []*astisub.Item{
			{
				Lines: []astisub.Line{{
					Items: []astisub.LineItem{{}},
				}},
			},
		},
		Regions: map[string]*astisub.Region{},
		Styles:  map[string]*astisub.Style{},
	}, s)
}

func TestSubtitles_ApplyLinearCorrection(t *testing.T) {
	s := &astisub.Subtitles{Items: []*astisub.Item{
		{
			EndAt:   2 * time.Second,
			StartAt: 1 * time.Second,
		},
		{
			EndAt:   5 * time.Second,
			StartAt: 3 * time.Second,
		},
		{
			EndAt:   10 * time.Second,
			StartAt: 7 * time.Second,
		},
	}}
	s.ApplyLinearCorrection(3*time.Second, 5*time.Second, 5*time.Second, 8*time.Second)
	require.Equal(t, 2*time.Second, s.Items[0].StartAt)
	require.Equal(t, 3500*time.Millisecond, s.Items[0].EndAt)
	require.Equal(t, 5*time.Second, s.Items[1].StartAt)
	require.Equal(t, 8*time.Second, s.Items[1].EndAt)
	require.Equal(t, 11*time.Second, s.Items[2].StartAt)
	require.Equal(t, 15500*time.Millisecond, s.Items[2].EndAt)
}
