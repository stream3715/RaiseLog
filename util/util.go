package util

import (
	"strconv"
)

//StrToInt64 ...Convert string to int64
func StrToInt64(s string, base int) (int64, error) {
	if base != 0 && (base < 2 || base > 36) {
		base = 10
	}
	return strconv.ParseInt(s, base, 64)
}
