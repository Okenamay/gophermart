package luhn

import "strconv"

// IsValid проверяет корректность номера по алгоритму Луна.
func IsValid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 { // even
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		number /= 10
	}
	return luhn % 10
}

// IsValidString является обёрткой для IsValid для строковых номеров.
func IsValidString(numStr string) bool {
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return false
	}
	return IsValid(num)
}
