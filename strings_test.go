package wflambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrings(t *testing.T) {
	assert := assert.New(t)

	b := stringToBool("true")
	assert.True(*b)
	b = stringToBool("false")
	assert.False(*b)
	b = stringToBool("bla")
	assert.False(*b)

	i, err := stringToInt("12")
	assert.Equal(*i, 12)
	assert.NoError(err)
	_, err = stringToInt("bla")
	assert.Error(err)
}
