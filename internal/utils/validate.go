package utils

import (
	"strings"
	"time"
)

func PrepareString(input string) string {
	trimmed := strings.TrimSpace(input)

	return trimmed
}

func PrepareStringPointer(input *string) *string {
	if input != nil {
		trimmed := strings.TrimSpace(*input)
		return &trimmed
	}

	return nil
}

func MustParseTime(ts string) time.Time {
	t, err := time.Parse(time.DateOnly, ts)
	if err != nil {
		panic(err.Error())
	}
	return t
}
