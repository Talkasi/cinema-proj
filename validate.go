package main

import (
	"strings"
	"time"
)

// PrepareString безопасно подготавливает строку для работы с БД
func PrepareString(input string) string {
	trimmed := strings.TrimSpace(input)

	return trimmed
}

func MustParseTime(ts string) time.Time {
	t, err := time.Parse(time.DateOnly, ts)
	if err != nil {
		panic(err.Error())
	}
	return t
}
