package luhn

import "strconv"

// IsValid checks the validity of a number using the Luhn algorithm.
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

// IsValidString is a wrapper for IsValid to check string numbers.
func IsValidString(numStr string) bool {
	if _, err := strconv.Atoi(numStr); err != nil {
		return false
	}

	var sum int
	nDigits := len(numStr)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit, _ := strconv.Atoi(string(numStr[i]))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
