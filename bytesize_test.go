// Note to self: do not use gofmt() on this file except that thou proceedeth to
// use the elisp function in the final comment to make tryEveryCase() look nice
// again.

package bytesize

import (
	"fmt"
	"testing"
)

type metricPrefix byte

const (
	_  = metricPrefix(iota)
	p0 // suffix is "B" (or "" for "%.0v" format)
	pK // suffix is "KiB" or "KB" or "K"
	pM // like pK: suffix is "MiB" or "MB" or "M"
	pG // like pK
	pT // like pK
	pP // like pK
	pE // like pK
)

type tryCaseFunc func(t *testing.T, num int64, numstr string, code metricPrefix)

func tryEveryCase(t *testing.T, f tryCaseFunc) {
	// NOTE elisp code at end of this file

//	f(t, 1234, "X.YZ", pK) // For testing the test harness

	f(t,           0,        "0",  p0)
	f(t,           1,        "1",  p0)
	f(t,        1000,     "1000",  p0)
	f(t,        1023,     "1023",  p0)

	f(t,          -1,       "-1",  p0)

	f(t,       -1023,    "-1023",  p0)
	f(t,       -1024,       "-1",  pK)
	f(t,       -1025,    "-1.00",  pK)

	f(t,        1024,        "1",  pK)
	f(t,        1025,     "1.00",  pK)
	f(t,        1029,     "1.00",  pK)
	f(t,        1030,     "1.01",  pK)
	f(t,        1039,     "1.01",  pK)
	f(t,        1049,     "1.02",  pK)

	f(t,        5119,     "5.00",  pK)
	f(t,        5120,        "5",  pK)
	f(t,        5121,     "5.00",  pK)

	f(t,        9000,     "8.79",  pK)

	f(t,       10234,     "9.99",  pK)
	f(t,       10235,     "10.0",  pK)
	f(t,       10239,     "10.0",  pK)
	f(t,       10240,       "10",  pK)
	f(t,       10241,     "10.0",  pK)
	f(t,       10291,     "10.0",  pK)
	f(t,       10292,     "10.1",  pK)

	f(t,      102348,     "99.9",  pK)
	f(t,      102349,      "100",  pK)
	f(t,      102400,      "100",  pK)
	f(t,      102911,      "100",  pK)
	f(t,      102912,      "101",  pK)
	f(t,      102920,      "101",  pK)

	f(t,     1048063,     "1023",  pK)
	f(t,     1048064,     "1024",  pK)
	f(t,     1048575,     "1024",  pK)
	f(t,     1048576,        "1",  pM)
	f(t,     1048577,     "1.00",  pM)

	f(t,       1<<30,        "1",  pG)
	f(t,       1<<40,        "1",  pT)
	f(t,       1<<50,        "1",  pP)
	f(t,       1<<60,        "1",  pE)
	f(t,    1025<<50,     "1.00",  pE)
	f(t,    1256<<50,     "1.23",  pE)
	f(t,    8191<<50,     "8.00",  pE)
}

var letterForCode = map[metricPrefix]string{
	pK: "K", pM: "M", pG: "G", pT: "T", pP: "P", pE: "E",
}

func TestToString(t *testing.T) {
	tryEveryCase(t, func(t *testing.T, num int64, numstr string, code metricPrefix) {
		v := ByteSize(num)
		var expected string  // expected result of .String(), %v, %.3v etc
		var expected0 string // expected result of %.0v
		var expected1 string // expected result of %.1v
		var expected2 string // expected result of %.2v
		if code == p0 {
			expected = numstr + "B"
			expected0 = numstr
			expected1 = expected
			expected2 = expected
		} else {
			expected0 = numstr + letterForCode[code]
			expected1 = expected0
			expected2 = expected0 + "B"
			expected = expected0 + "iB"
		}
		// v.String()
		actual := v.String()
		if actual != expected {
			t.Errorf("Stringing %11d => %12q, wanted %12q",
				num, actual, expected)
		}
		// formatting with %v, %.0v, etc
		actual = fmt.Sprintf("%v", v)
		if actual != expected {
			tryFormat(t, "%v", v, expected) // log an error
		} else {
			tryFormat(t, "%.0v", v, expected0)
			tryFormat(t, "%.1v", v, expected1)
			tryFormat(t, "%.2v", v, expected2)
			tryFormat(t, "%.3v", v, expected)
			tryFormat(t, "%.4v", v, expected)
			tryFormat(t, "%.45678v", v, expected)
		}
	})
}

func tryFormat(t *testing.T, format string, v ByteSize, expected string) {
	actual := fmt.Sprintf(format, v)
	if actual != expected {
		t.Errorf("Formatting %10d with %6q => %12q, wanted %12q",
			v, format, actual, expected)
	}
}

/*
;;; Emacs lisp code to nicely format tryEveryCase() code
(defun fix-bytesize-test-code () (interactive "*")
    (save-match-data
      (goto-char (point-min))
      (while (re-search-forward
              "f(t, \\([-0-9<]+\\), \\(\"[-0-9.]+\"\\), +\\(p[0KMGTPE]\\))"
              nil t)
        (replace-match (format "f(t,%12s,  %9s,  %s)"
                               (match-string 1) (match-string 2) (match-string 3))
                       t))))
*/
