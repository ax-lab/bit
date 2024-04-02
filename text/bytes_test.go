package text_test

import (
	"testing"

	"axlab.dev/test/text"
	"github.com/stretchr/testify/require"
)

func TestBytes(t *testing.T) {
	test := require.New(t)

	test.Equal("1 byte", text.Bytes(1))
	test.Equal("0 bytes", text.Bytes(0))
	test.Equal("2 bytes", text.Bytes(2))
	test.Equal("42 bytes", text.Bytes(42))
	test.Equal("1023 bytes", text.Bytes(1023))

	test.Equal("1 kilobyte", text.Bytes(1024))
	test.Equal("1 kilobyte", text.Bytes(1025))
	test.Equal("2 kilobytes", text.Bytes(1999))
	test.Equal("1023 kilobytes", text.Bytes(1023*text.KB))
	test.Equal("1 megabyte", text.Bytes(1024*text.KB))
	test.Equal("1.2 megabytes", text.Bytes(1.2*text.MB))
	test.Equal("1.25 megabytes", text.Bytes(1.25*text.MB))

	var mult float64

	mult = 1
	test.Equal("1 kilobyte", text.Bytes(mult*1024))
	test.Equal("1 megabyte", text.Bytes(mult*1024*1024))
	test.Equal("1 gigabyte", text.Bytes(mult*1024*1024*1024))
	test.Equal("1 terabyte", text.Bytes(mult*1024*1024*1024*1024))

	mult = 1.2
	test.Equal("1 kilobyte", text.Bytes(mult*1024))
	test.Equal("1.2 megabytes", text.Bytes(mult*1024*1024))
	test.Equal("1.2 gigabytes", text.Bytes(mult*1024*1024*1024))
	test.Equal("1.2 terabytes", text.Bytes(mult*1024*1024*1024*1024))

	mult = 1.25
	test.Equal("1 kilobyte", text.Bytes(mult*1024))
	test.Equal("1.25 megabytes", text.Bytes(mult*1024*1024))
	test.Equal("1.25 gigabytes", text.Bytes(mult*1024*1024*1024))
	test.Equal("1.25 terabytes", text.Bytes(mult*1024*1024*1024*1024))

	mult = 1.9
	test.Equal("2 kilobytes", text.Bytes(mult*1024))
	test.Equal("1.9 megabytes", text.Bytes(mult*1024*1024))
	test.Equal("1.9 gigabytes", text.Bytes(mult*1024*1024*1024))
	test.Equal("1.9 terabytes", text.Bytes(mult*1024*1024*1024*1024))

	mult = 1.125
	test.Equal("1 kilobyte", text.Bytes(mult*1024))
	test.Equal("1.13 megabytes", text.Bytes(mult*1024*1024))
	test.Equal("1.13 gigabytes", text.Bytes(mult*1024*1024*1024))
	test.Equal("1.125 terabytes", text.Bytes(mult*1024*1024*1024*1024))
}

func TestBytesShort(t *testing.T) {
	test := require.New(t)

	test.Equal("1B", text.BytesShort(1))
	test.Equal("0B", text.BytesShort(0))
	test.Equal("2B", text.BytesShort(2))
	test.Equal("42B", text.BytesShort(42))
	test.Equal("1023B", text.BytesShort(1023))

	test.Equal("1KB", text.BytesShort(1024))
	test.Equal("1KB", text.BytesShort(1025))
	test.Equal("2KB", text.BytesShort(1999))
	test.Equal("1023KB", text.BytesShort(1023*text.KB))
	test.Equal("1MB", text.BytesShort(1024*text.KB))
	test.Equal("1.2MB", text.BytesShort(1.2*text.MB))
	test.Equal("1.25MB", text.BytesShort(1.25*text.MB))

	var mult float64

	mult = 1
	test.Equal("1KB", text.BytesShort(mult*1024))
	test.Equal("1MB", text.BytesShort(mult*1024*1024))
	test.Equal("1GB", text.BytesShort(mult*1024*1024*1024))
	test.Equal("1TB", text.BytesShort(mult*1024*1024*1024*1024))

	mult = 1.2
	test.Equal("1KB", text.BytesShort(mult*1024))
	test.Equal("1.2MB", text.BytesShort(mult*1024*1024))
	test.Equal("1.2GB", text.BytesShort(mult*1024*1024*1024))
	test.Equal("1.2TB", text.BytesShort(mult*1024*1024*1024*1024))

	mult = 1.25
	test.Equal("1KB", text.BytesShort(mult*1024))
	test.Equal("1.25MB", text.BytesShort(mult*1024*1024))
	test.Equal("1.25GB", text.BytesShort(mult*1024*1024*1024))
	test.Equal("1.25TB", text.BytesShort(mult*1024*1024*1024*1024))

	mult = 1.9
	test.Equal("2KB", text.BytesShort(mult*1024))
	test.Equal("1.9MB", text.BytesShort(mult*1024*1024))
	test.Equal("1.9GB", text.BytesShort(mult*1024*1024*1024))
	test.Equal("1.9TB", text.BytesShort(mult*1024*1024*1024*1024))

	mult = 1.125
	test.Equal("1KB", text.BytesShort(mult*1024))
	test.Equal("1.13MB", text.BytesShort(mult*1024*1024))
	test.Equal("1.13GB", text.BytesShort(mult*1024*1024*1024))
	test.Equal("1.125TB", text.BytesShort(mult*1024*1024*1024*1024))
}
