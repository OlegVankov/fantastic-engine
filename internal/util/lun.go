package util

import "strconv"

func CheckLun(num string) bool {
	sum := 0
	parity := len(num) % 2

	for i, v := range num {
		digit, _ := strconv.Atoi(string(v))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
