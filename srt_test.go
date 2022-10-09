package astisub_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/molotovtv/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestSRT(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.srt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSRT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.srt")
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestSRTWithStyling(t *testing.T) {
	testData := `1
00:00:08,543 --> 00:00:13,143
BLUE HOLOCAUST
<i>BUIO OMEGA</i>

2
00:02:14,267 --> 00:02:16,136
<b>Quand tu l'auras fini,</b>
je viendrai le chercher.

3
00:02:16,194 --> 00:02:18,062
- Au revoir !
- À plus tard.

4
00:04:11,251 --> 00:04:12,414
<u>Maman...</u>

5
00:04:14,213 --> 00:04:15,921
- Maman...
- Anna !

6
00:04:17,425 --> 00:04:19,336
<i>J’étouffe .</i>
`

	s, err := astisub.ReadFromSRT(strings.NewReader(testData))
	assert.NoError(t, err)

	b := &bytes.Buffer{}
	err = s.WriteToSRT(b)
	assert.NoError(t, err)
	assert.Equal(t, string(append(astisub.BytesBOM, []byte(`1
00:00:08,543 --> 00:00:13,143
BLUE HOLOCAUST
<i>BUIO OMEGA</i>

2
00:02:14,267 --> 00:02:16,136
<b>Quand tu l'auras fini,</b>
je viendrai le chercher.

3
00:02:16,194 --> 00:02:18,062
- Au revoir !
- À plus tard.

4
00:04:11,251 --> 00:04:12,414
<u>Maman...</u>

5
00:04:14,213 --> 00:04:15,921
- Maman...
- Anna !

6
00:04:17,425 --> 00:04:19,336
<i>J’étouffe .</i>
`)...)), b.String())
}
