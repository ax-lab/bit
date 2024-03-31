package text_test

import (
	"strings"
	"testing"

	"axlab.dev/test/text"
	"github.com/stretchr/testify/require"
)

func TestScanRead(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := text.ScannerNew(INPUT)

	for _, expected := range INPUT {
		chr, read := scan.Read()
		test.True(read)
		test.Equal(expected, chr)
	}

	_, read := scan.Read()
	test.False(read)
}

func TestScanPeek(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := text.ScannerNew(INPUT)

	for _, expected := range INPUT {
		test.Equal(expected, scan.Peek())
		scan.Read()
	}
	test.Equal(rune(0), scan.Peek())
}

func TestScanText(t *testing.T) {
	test := require.New(t)

	const INPUT = "abc123"
	scan := text.ScannerNew(INPUT)

	for n := range INPUT {
		rest := INPUT[n:]
		test.Equal(rest, scan.Text())
		test.Equal(len(rest), scan.Len())
		scan.Read()
	}

	test.Equal("", scan.Text())
	test.Equal(0, scan.Len())
}

func TestScanReadChars(t *testing.T) {
	test := require.New(t)

	const INPUT = "a12abc1234"
	scan := text.ScannerNew(INPUT)

	test.Equal("", scan.ReadChars(0))
	test.Equal("a", scan.ReadChars(1))
	test.Equal("12", scan.ReadChars(2))
	test.Equal("abc", scan.ReadChars(3))
	test.Equal("1234", scan.ReadChars(4))
	test.Equal("", scan.ReadChars(999))
}

func TestScanReadWhile(t *testing.T) {
	test := require.New(t)

	const INPUT = "12345ABCDE"
	scan := text.ScannerNew(INPUT)

	mT := text.MatchRegex(`\d`)
	mF := text.MatchRegex(`[^\d]`)
	test.Equal("", scan.ReadWhile(mF))
	test.Equal("12345", scan.ReadWhile(mT))
	test.Equal("ABCDE", scan.ReadWhile(mF))
}

func TestScanReadUntil(t *testing.T) {
	test := require.New(t)

	const INPUT = "12345ABCDE"
	scan := text.ScannerNew(INPUT)

	mT := text.MatchRegex(`\d`)
	mF := text.MatchRegex(`[^\d]`)
	test.Equal("", scan.ReadUntil(mT))
	test.Equal("12345", scan.ReadUntil(mF))
	test.Equal("ABCDE", scan.ReadUntil(mT))
}

func TestScanReadMatch(t *testing.T) {
	test := require.New(t)

	const INPUT = "42a123b456c"
	scan := text.ScannerNew(INPUT)

	var pre, txt string
	m := text.MatchRegex(`\d+`)

	pre, txt = scan.ReadMatch(m)
	test.Equal("", pre)
	test.Equal("42", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("a", pre)
	test.Equal("123", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("b", pre)
	test.Equal("456", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("c", pre)
	test.Equal("", txt)

	pre, txt = scan.ReadMatch(m)
	test.Equal("", pre)
	test.Equal("", txt)
}

var benchInput = (func() string {
	const textSize = 1024 * 1024
	out := strings.Builder{}
	for n := 0; n <= textSize; n++ {
		chr := rune(n % 16)
		if chr >= 0x0A {
			chr += 'A'
		} else {
			chr += '0'
		}
		out.WriteRune(chr)
	}
	return out.String()
}())

func BenchmarkRawRead(t *testing.B) {
	for n := 0; n < t.N; n++ {
		sum := int32(0)
		for _, chr := range benchInput {
			sum += chr
		}
	}

	elapsed := t.Elapsed().Seconds()
	bytes := float64(t.N) * float64(len(benchInput)) / (1024 * 1024)
	t.ReportMetric(bytes/elapsed, "MB/s")
}

func BenchmarkScanRead(t *testing.B) {
	for n := 0; n < t.N; n++ {
		sum := int32(0)
		scan := text.ScannerNew(benchInput)
		for {
			if chr, ok := scan.Read(); ok {
				sum += chr
			} else {
				break
			}
		}
	}

	elapsed := t.Elapsed().Seconds()
	bytes := float64(t.N) * float64(len(benchInput)) / (1024 * 1024)
	t.ReportMetric(bytes/elapsed, "MB/s")
}

func BenchmarkScanMatchRegex(t *testing.B) {
	matched := 0
	matchDigit := text.MatchRegex(`\d+`)
	t.ResetTimer()

	for n := 0; n < t.N; n++ {
		scan := text.ScannerNew(benchInput)
		for scan.Len() > 0 {
			_, digits := scan.ReadMatch(matchDigit)
			matched += len(digits)
		}
	}

	elapsed := t.Elapsed().Seconds()
	total := float64(t.N) * float64(len(benchInput))
	bytes := total / (1024 * 1024)
	t.ReportMetric(bytes/elapsed, "MB/s")

	t.ReportMetric(100*float64(matched)/total, "%Match")
}

func BenchmarkScanMatchIf(t *testing.B) {
	matchDigit := text.MatchIf(func(chr rune) bool { return '0' <= chr && chr <= '9' })
	matched := 0
	for n := 0; n < t.N; n++ {
		scan := text.ScannerNew(benchInput)
		for scan.Len() > 0 {
			_, digits := scan.ReadMatch(matchDigit)
			matched += len(digits)
		}
	}

	elapsed := t.Elapsed().Seconds()
	total := float64(t.N) * float64(len(benchInput))
	bytes := total / (1024 * 1024)
	t.ReportMetric(bytes/elapsed, "MB/s")

	t.ReportMetric(100*float64(matched)/total, "M%")
}
