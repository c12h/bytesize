// Package bytesize provides type ByteSize, a number of bytes, which prints as
// "1B", "2KiB", "3.45MiB", "67.8GiB", "987TiB", "1023EiB" etc.
//
// ByteSize is derived from int64, so int64(x) returns the underlying value of x.
//
// Note that ByteSize values may be negative.
//
// When printing a ByteSize value with a format string, the %s and %v verbs
// print the value with 1 to 4 decimal digits, possibly including a decimal
// point, followed by "B", "KiB", "MiB", "GiB", "TiB", "PiB" or "EiB" as
// appropriate.  All other verbs, notably %d, act as though applied to the
// underlying number.
//
// Specifying a precision in a format specifier does not affect how many digits
// are output. Instead, a precision can select shorter forms of the suffixes:
//  Precision:	Suffixes:-
//	.0	"",    "K",   "M",   "G",   "T",   "P",   "E"
//	.1	"B",   "K",   "M",   "G",   "T",   "P",   "E"
//	.2	"B",   "KB",  "MB",  "GB",  "TB",  "PB",  "EB"
//	.3	"B",   "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"
// Other precision values are treated as .3, giving the default suffixes.
//
// ByteSize has a String() method, which always uses the default suffixes.
//
// For completeness, ByteSize also has a GoString method, which has the same
// effect as fmt.Sprintf("ByteSize(%d)", int64(value)).
//
// This package exports only one directly-visible name, ByteSize.
//
// When using this package, you may want to define a type alias, like this:
//	type ByteSize = bytesize.ByteSize
// so you can write "ByteSize" instead of "bytesize.ByteSize".
//
package bytesize

// PROPOSAL: Should internationalize this		// ???FIXME
// PROPOSAL: could also implement fmt.Scanner

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// A ByteSize is a number of bytes, possibly negative.
//
type ByteSize int64

// The String() method returns what formatting the value with "%.3v" would:
// strings like "0B", ..., "1023B", "1KiB", ..., "1.01Kib", "-23.4MiB", "340GiB"
// etc.
//
// This method makes ByteSize satisfy the fmt.Stringer interface.  (Because
// ByteSize also satisfies the fmt.Formatter interface, fmt will use .Format(…)
// rather than this method when printing ByteSize values.)
//
func (n ByteSize) String() string {
	return string(formatByteSize(int64(n), 3))
}

// The GoString() method is equivalent to formatting the underlying int64 value
// with "ByteSize(%d)".
//
// This method makes ByteSize satisfy the fmt.GoStringer interface.  (Because
// ByteSize also satisfies the fmt.Formatter interface, fmt will use .Format(…)
// rather than this method when printing ByteSize values.)
//
func (n ByteSize) GoString() string {
	return fmt.Sprintf("ByteSize(%d)", int64(n))
}

// The Format method makes ByteSize satisfy the fmt.Formatter interface, and thus
// controls how ByteSize values are printed.
//
// For verbs other than %s and %v, it simply applies the same conversion specifier
// to the underlying int64 value.
//
// For %#v, it outputs fmt.Sprintf("ByteSize(%d)", int64(n)).
//
// For %s and %v without "#", it outputs a decimal number with 3 or 4
// significant digits followed by the appropriate suffix.  The output may or may
// not have a decimal point.  As mentioned in the package documentation, a
// precision specifier format verb affects which suffixes are used, not how many
// digits are output.
//
func (b ByteSize) Format(f fmt.State, verb rune) {
	nBytes := int64(b)
	if verb == 'v' {
		if f.Flag('#') {
			fmt.Fprintf(f, "ByteSize(%d)", nBytes)
			return
		}
		// else use formatByteSize()
	} else if verb != 's' {
		fmt.Fprintf(f, equivalentFormat(f, verb), nBytes)
		return
	}

	prec, specified := f.Precision()
	if !specified || prec > 3 {
		prec = 3
	} else if prec < 0 {
		prec = 0
	}

	output := formatByteSize(nBytes, prec)

	width, haveWidth := f.Width()
	if !haveWidth {
		f.Write(output)
	} else {
		diff := width - utf8.RuneCount(output)
		padding := bytes.Repeat([]byte(" "), diff)
		if f.Flag('-') {
			f.Write(output)
			f.Write(padding)
		} else {
			f.Write(padding)
			f.Write(output)
		}
	}
}

// equivalentFormat() constructs a format specifier string equivalent to the one
// represented by a fmt.State value.
func equivalentFormat(f fmt.State, verb rune) string {
	formatString := "%"
	for _, ch := range []rune{'#', '+', '-', ' ', '0'} {
		if f.Flag(int(ch)) {
			formatString += string(ch)
		}
	}
	if w, present := f.Width(); present {
		formatString += strconv.FormatInt(int64(w), 10)
	}
	if p, present := f.Precision(); present {
		formatString += "."
		formatString += strconv.FormatInt(int64(p), 10)
	}
	return formatString + string(verb)
}

// formatByteSize does the hard work for this package.
func formatByteSize(value int64, prec int) []byte {
	const suffix1 = "!KMGTPE"
	const _1 = int64(1)
	ret := make([]byte, 0, 32) // Plenty of room.
	letter := byte(0)
	if value < 0 {
		ret = append(ret, '-')
		value = -value
	}
	if value < 1024 {
		// Values < 1024 get different suffixes.
		ret = appendDecimal(ret, value)
		if prec != 0 {
			ret = append(ret, 'B')
		}
		return ret
	}
	if value >= (1 << 60) {
		// This is a simple way to avoid integer overflows
		const unitSize = 1 << 60
		if (value & (unitSize - 1)) == 0 {
			// Exact multiples of unitSize are a special case.
			ret = appendDecimal(ret, value>>60)
		} else {
			ret = appendDecimal(ret, ((value>>50)*100+512)>>10)
			//D// hook1("EiB, decimals==2, ret=%q", ret)
			n := len(ret)
			ret = append(ret[:n-2], '.', ret[n-2], ret[n-1])
		}
		letter = 'E'
	} else {
		level := uint(1)
		for k := (_1 << 20); value >= k && k < (_1<<60); k <<= 10 {
			level++
		}
		unitSize := int64(1 << (10 * level)) // 2**10 or 2**20 or ... or 2**60
		// Now 1 <= level <= 6 and unitSize <= value <= 1023*unitSize

		// Format value÷unitSize in decimal with 1-4 digits.
		if (value & (unitSize - 1)) == 0 {
			// Exact multiples of unitSize are a special case.
			ret = appendDecimal(ret, value>>(10*level))
		} else {
			// For other values, round up in the last digit.
			digitsAfterPoint, v := 0, value
			if value < 10*unitSize-unitSize/200 {
				digitsAfterPoint = 2
				v = 100 * value
			} else if value < 100*unitSize-unitSize/20 {
				digitsAfterPoint = 1
				v = 10 * value
			}
			ret = appendDecimal(ret, (v+unitSize/2)>>(10*level))
			// Maybe insert a decimal point.
			n := len(ret)
			switch digitsAfterPoint {
			case 2:
				ret = append(ret[:n-2], '.', ret[n-2], ret[n-1])
			case 1:
				ret = append(ret[:n-1], '.', ret[n-1])
			}
		}
		letter = suffix1[level]
	}
	switch prec {
	case 0, 1:
		ret = append(ret, letter)
	case 2:
		ret = append(ret, letter, 'B')
	default:
		ret = append(ret, letter, 'i', 'B')
	}
	return ret
}

// appendDecimal() appends the decimal form of an int64 value, which MUST be
// positive, to a byte slice.
//
func appendDecimal(buffer []byte, number int64) []byte {
	var buf [20]byte // big enough for 64bit value base 10
	i := len(buf) - 1
	for number >= 10 {
		q := number / 10
		buf[i] = byte('0' + number - q*10)
		i--
		number = q
	}
	buf[i] = byte('0' + number)
	return append(buffer, buf[i:]...)
}
