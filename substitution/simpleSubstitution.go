package substitution

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"time"
)

type parseData struct {
	p int    // position
	c byte   // current char
	b []byte // Text as bytes
	l int    // langth of Text as bytes
}

/*
DoSubstitution - replace ${n} in string from:
map passed in
environment variables
date and time
*/
func DoSubstitution(text string, dataIn map[string]string, tag byte) string {
	if text == "" {
		return text
	}
	by := []byte(text)
	pa := &parseData{
		p: -1,
		b: by,
		c: 0,
		l: len(by),
	}
	var buf bytes.Buffer
	for pa.next() {
		if pa.c == tag {
			if pa.next() {
				if pa.c == '{' {
					n, f := pa.upTo('}')
					s := ""
					if dataIn != nil {
						s = dataIn[n]
					}
					if s == "" {
						s = systemSubs(n)
					}
					if s == "" {
						buf.WriteByte(tag)
						if f {
							buf.WriteString("{" + n + "}")
						} else {
							buf.WriteString("{" + n)
						}
					} else {
						buf.WriteString(s)
					}
				} else {
					buf.WriteByte(tag)
					buf.WriteByte(pa.c)
				}
			} else {
				buf.WriteByte(tag)
			}
		} else {
			buf.WriteByte(pa.c)
		}
	}
	return buf.String()
}

func (p *parseData) next() bool {
	p.p++
	if p.p < p.l {
		p.c = p.b[p.p]
		return true
	}
	return false
}

func (p *parseData) upTo(c byte) (string, bool) {
	var buf bytes.Buffer
	for p.next() {
		if p.c != c {
			buf.WriteByte(p.c)
		} else {
			return buf.String(), true
		}
	}
	return buf.String(), false
}

func systemSubs(n string) string {
	tim := time.Now()
	yyyy, mm, dd := tim.Date()
	switch n {
	case "YYYY":
		return padString2(yyyy)
	case "MM":
		return padString2(int(mm))
	case "DD":
		return padString2(int(dd))
	case "HH":
		return padString2(tim.Hour())
	case "mm":
		return padString2(tim.Minute())
	case "SS":
		return padString2(tim.Second())
	case "PID":
		return padString2(os.Getpid())
	case "OS":
		return runtime.GOOS
	}
	return os.Getenv(n)
}

func padString2(i int) string {
	if i < 10 {
		return "0" + strconv.Itoa(i)
	}
	return strconv.Itoa(i)
}
