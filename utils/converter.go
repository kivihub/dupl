package utils

import (
	"math"
	"strconv"
	"strings"
)

func ParseStorageToBytes(storageStr string) int64 {
	if storageStr == "" {
		return math.MaxInt64
	}

	unitMultiplier := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	storageStr = strings.ToUpper(strings.TrimSpace(storageStr))
	var numberStr string
	var unit string

	for u := range unitMultiplier {
		if strings.HasSuffix(storageStr, u) {
			unit = u
			numberStr = strings.TrimSuffix(storageStr, u)
			break
		}
	}
	if unit == "" && strings.HasSuffix(storageStr, "B") {
		unit = "B"
		numberStr = strings.TrimSuffix(storageStr, unit)
	}

	if unit == "" {
		unit = "B"
		numberStr = storageStr
	}

	num, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return math.MaxInt64
	}

	return int64(num) * unitMultiplier[unit]
}
