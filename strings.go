package wflambda

import "strconv"

func stringToBool(s string) *bool {
	b := false
	if s == "true" {
		b = true
	}
	return &b
}

func stringToInt(s string) (*int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
