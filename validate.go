package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SafeStringParams параметры для настройки обработки строки
type SafeStringParams struct {
	MaxLength   int            // Максимальная длина (0 - без ограничения)
	AllowEmpty  bool           // Разрешать пустые строки
	CustomRegex *regexp.Regexp // regex-правило
}

// PrepareString безопасно подготавливает строку для работы с БД
func PrepareString(input string, params *SafeStringParams) (string, error) {
	trimmed := strings.TrimSpace(input)

	if params != nil {
		if !params.AllowEmpty && trimmed == "" {
			return "", ErrEmptyString
		}

		if params.MaxLength > 0 && utf8.RuneCountInString(trimmed) > params.MaxLength {
			return "", fmt.Errorf("допустимый размер строки %v: %v", params.MaxLength, ErrTooBigString.Error())
		}

		if params.CustomRegex != nil && !params.CustomRegex.MatchString(trimmed) {
			return "", ErrInvalidChars
		}
	}

	// escaped := escapeSQL(trimmed)
	return trimmed, nil
}

// escapeSQL экранирование специальных символов (использовать только в крайних случаях!)
func escapeSQL(input string) string {
	var escape = map[rune]string{
		'\'':   "''",
		'"':    `\"`,
		'\\':   `\\`,
		'\x00': "\\x00",
		'\n':   "\\n",
		'\r':   "\\r",
	}

	var builder strings.Builder
	for _, char := range input {
		if escaped, ok := escape[char]; ok {
			builder.WriteString(escaped)
		} else {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

// Ошибки валидации
var (
	ErrEmptyString  = errors.New("пустая строка не допускается")
	ErrInvalidChars = errors.New("строка содержит недопустимые символы")
	ErrTooBigString = errors.New("размер строки превышает допустимый")
)
