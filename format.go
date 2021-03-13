package timefmt

import (
	"strconv"
	"time"
)

// Format time to string using the format.
func Format(t time.Time, format string) string {
	return string(AppendFormat(make([]byte, 0, 64), t, format))
}

// AppendFormat appends formatted time to the bytes.
// You can use this method to reduce allocations.
func AppendFormat(buf []byte, t time.Time, format string) []byte {
	year, month, day := t.Date()
	var width int
	var padding byte
	var pending string
	for i := 0; i < len(format); i++ {
		if b := format[i]; b == '%' {
			if i++; i == len(format) {
				buf = append(buf, '%')
				break
			}
			b, width, padding = format[i], 0, '0'
		L:
			switch b {
			case '-':
				if pending != "" {
					buf = append(buf, '-')
					break
				}
				if i++; i == len(format) {
					goto K
				}
				padding = ^paddingMask
				b = format[i]
				goto L

			case 'Y':
				if width == 0 {
					width = 4
				}
				buf = appendInt(buf, year, width, padding)

			case 'm':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, int(month), width, padding)

			case 'd':
				if width < 2 {
					width = 2
				}
				buf = appendInt(buf, day, width, padding)
			case '%':
				buf = appendString(buf, "%", width, padding, false, false)
			default:
				if pending == "" {
					var ok bool
					if pending, ok = compositions[b]; ok {
						break
					}
					buf = appendLast(buf, format[:i], width-1, padding)
				}
				buf = append(buf, b)
			}
			if pending != "" {
				b, pending, width, padding = pending[0], pending[1:], 0, '0'
				goto L
			}
		} else {
			buf = append(buf, b)
		}
	}
	return buf
K:
	return appendLast(buf, format, width, padding)
}

func appendInt(buf []byte, num, width int, padding byte) []byte {
	if padding != ^paddingMask {
		padding &= paddingMask
		switch width {
		case 2:
			if num < 10 {
				buf = append(buf, padding)
				goto L1
			} else if num < 100 {
				goto L2
			} else if num < 1000 {
				goto L3
			} else if num < 10000 {
				goto L4
			}
		case 4:
			if num < 1000 {
				buf = append(buf, padding)
				if num < 100 {
					buf = append(buf, padding)
					if num < 10 {
						buf = append(buf, padding)
						goto L1
					}
					goto L2
				}
				goto L3
			} else if num < 10000 {
				goto L4
			}
		default:
			i := len(buf)
			for ; width > 1; width-- {
				buf = append(buf, padding)
			}
			j := len(buf)
			buf = strconv.AppendInt(buf, int64(num), 10)
			l := len(buf)
			if j+1 == l || i == j {
				return buf
			}
			k := j + 1 - (l - j)
			if k < i {
				l = j + 1 + i - k
				k = i
			} else {
				l = j + 1
			}
			copy(buf[k:], buf[j:])
			return buf[:l]
		}
	}
	if num < 100 {
		if num < 10 {
			goto L1
		}
		goto L2
	} else if num < 10000 {
		if num < 1000 {
			goto L3
		}
		goto L4
	}
	return strconv.AppendInt(buf, int64(num), 10)
L4:
	buf = append(buf, byte(num/1000)|'0')
	num %= 1000
L3:
	buf = append(buf, byte(num/100)|'0')
	num %= 100
L2:
	buf = append(buf, byte(num/10)|'0')
	num %= 10
L1:
	return append(buf, byte(num)|'0')
}

func appendString(buf []byte, str string, width int, padding byte, upper, swap bool) []byte {
	if width > len(str) && padding != ^paddingMask {
		if padding < ^paddingMask {
			padding = ' '
		} else {
			padding &= paddingMask
		}
		for width -= len(str); width > 0; width-- {
			buf = append(buf, padding)
		}
	}
	switch {
	case swap:
		if str[len(str)-1] < 'a' {
			for _, b := range []byte(str) {
				buf = append(buf, b|0x20)
			}
			break
		}
		fallthrough
	case upper:
		for _, b := range []byte(str) {
			buf = append(buf, b&0x5F)
		}
	default:
		buf = append(buf, str...)
	}
	return buf
}

func appendLast(buf []byte, format string, width int, padding byte) []byte {
	for i := len(format) - 1; i >= 0; i-- {
		if format[i] == '%' {
			buf = appendString(buf, format[i:], width, padding, false, false)
			break
		}
	}
	return buf
}

const paddingMask byte = 0x7F

var longMonthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longWeekNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortWeekNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var compositions = map[byte]string{
	'c': "a b e H:M:S Y",
	'+': "a b e H:M:S Z Y",
	'F': "Y-m-d",
	'D': "m/d/y",
	'x': "m/d/y",
	'v': "e-b-Y",
	'T': "H:M:S",
	'X': "H:M:S",
	'r': "I:M:S p",
	'R': "H:M",
}
