package utils

import (
	"fmt"
	"unicode"
)

func LowercaseMessage(message string) string {
	if len(message) >= 2 {
		firstChar := rune(message[0])
		secondChar := rune(message[1])

		if !unicode.IsLower(firstChar) && unicode.IsLower(secondChar) {
			return string(unicode.ToLower(firstChar)) + message[1:]
		}
	}
	return message
}

func Pluralize(count int64, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}

func RemoveTrailingPeriod(message string) string {
	if len(message) > 0 && message[len(message)-1] == '.' {
		return message[:len(message)-1]
	}
	return message
}

func TextFromSentence(message string) string {
	return RemoveTrailingPeriod(LowercaseMessage(message))
}
