package util

import (
	"strconv"
	"strings"
)

func ParseStrToInt64(p string) (*int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
	if err != nil {
		return nil, err
	}

	return &id, nil
}
