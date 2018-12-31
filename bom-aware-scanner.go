package astisub

import (
	"bufio"
	"bytes"
	"io"
)

type BomAwareScanner struct {
	*bufio.Scanner
	lookingForBom bool
}

func NewBomAwareScanner(r io.Reader) *BomAwareScanner {
	bas := BomAwareScanner{Scanner: bufio.NewScanner(r), lookingForBom: true}
	bas.Scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		switch {
		case bas.lookingForBom:
			if len(data) < len(BytesBOM) {
				if atEOF {
					bas.lookingForBom = false
					break
				}
				return 0, nil, nil
			}
			lowIndex := func()int{if bytes.HasPrefix(data, BytesBOM) {return len(BytesBOM)}; return 0}()
			bas.lookingForBom = false
			advance, token, err = bufio.ScanLines(data[lowIndex:], atEOF)
			advance += lowIndex
			return
		}
		return bufio.ScanLines(data, atEOF)
	})
	return &bas
}
