package wflambda

import (
	"strconv"
	"strings"
)

// stringToBool interprets a string s and returns a pointer to the corresponding boolean value i.
// A pointer to a true boolean is returned when the input string is true, a false boolean is returned
// in all other cases. The string match is case-insensitive.
func stringToBool(s string) *bool {
	b := false
	if strings.EqualFold(s, "true") {
		b = true
	}
	return &b
}

// stringToInt interprets a string s in base 10 format and bit size 0 and returns a pointer to the
// corresponding value i. An error is returned when converting the string to an integer fails.
func stringToInt(s string) (*int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
