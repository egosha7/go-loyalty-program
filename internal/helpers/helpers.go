package helpers

import "strings"

func IsLuhnValid(number string) bool {
	// Алгоритм Луна для проверки корректности номера
	number = strings.Replace(number, " ", "", -1)
	if len(number) < 2 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
