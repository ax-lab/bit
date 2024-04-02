package text

import (
	"fmt"
	"math"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

type Numeric interface {
	~int | ~int64 | ~int32 | ~int16 | ~int8 | ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8 | ~float32 | ~float64
}

func Bytes[T Numeric](value T) string {
	return getBytes(value, false)
}

func BytesShort[T Numeric](value T) string {
	return getBytes(value, true)
}

func getBytes[T Numeric](value T, short bool) string {
	if value < 0 {
		return "-" + getBytes(-value, short)
	}

	var (
		unitShort = [5]string{"B", "KB", "MB", "GB", "TB"}
		unitLong  = [5]string{" byte", " kilobyte", " megabyte", " gigabyte", " terabyte"}
	)

	units := unitLong
	if short {
		units = unitShort
	}

	var (
		div    float64
		unit   string
		places int
	)

	bytes := float64(value)
	if bytes < KB {
		div, unit, places = 1, units[0], 0
	} else if bytes < MB {
		div, unit, places = KB, units[1], 0
	} else if bytes < GB {
		div, unit, places = MB, units[2], 2
	} else if bytes < TB {
		div, unit, places = GB, units[3], 2
	} else {
		div, unit, places = TB, units[4], 3
	}

	output, round := bytes/div, math.Pow10(places)
	output = math.Round(output * round)

	for output >= 10 && round >= 10 && int64(output)%10 == 0 {
		output /= 10
		round /= 10
		places -= 1
	}
	output = output / round

	plural := ""
	if output != 1 && unit[len(unit)-1] != 'B' {
		plural = "s"
	}

	return fmt.Sprintf("%.*f%s%s", places, output, unit, plural)
}
