package valueobject

import (
	"fmt"
	"regexp"
	"strconv"
	"usuario-service/internal/domain/errors"
)

type CPF struct {
	value string
}

var nonDigitRegex = regexp.MustCompile(`[^\d]`)

func NewCPF(value string) (CPF, error) {
	digits := nonDigitRegex.ReplaceAllString(value, "")

	if len(digits) != 11 {
		return CPF{}, domainerrors.ErrCPFInvalido
	}

	// Check if all digits are the same
	allSame := true
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return CPF{}, domainerrors.ErrCPFInvalido
	}

	// Check first digit
	sum1 := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(digits[i]))
		sum1 += digit * (10 - i)
	}
	rem1 := sum1 % 11
	d1 := 0
	if rem1 >= 2 {
		d1 = 11 - rem1
	}
	if fmt.Sprintf("%d", d1) != string(digits[9]) {
		return CPF{}, domainerrors.ErrCPFInvalido
	}

	// Check second digit
	sum2 := 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(digits[i]))
		sum2 += digit * (11 - i)
	}
	rem2 := sum2 % 11
	d2 := 0
	if rem2 >= 2 {
		d2 = 11 - rem2
	}
	if fmt.Sprintf("%d", d2) != string(digits[10]) {
		return CPF{}, domainerrors.ErrCPFInvalido
	}

	formatted := fmt.Sprintf("%s.%s.%s-%s", digits[0:3], digits[3:6], digits[6:9], digits[9:11])
	return CPF{value: formatted}, nil
}

func (c CPF) Value() string {
	return c.value
}

func (c CPF) Unformatted() string {
	return nonDigitRegex.ReplaceAllString(c.value, "")
}

func (c CPF) Equals(other CPF) bool {
	return c.value == other.value
}

func (c CPF) String() string {
	return c.value
}
