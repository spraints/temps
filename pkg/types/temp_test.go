package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversions(t *testing.T) {
	assert.Equal(t, Celsius(0), Fahrenheit(32).Celsius())
	assert.Equal(t, Fahrenheit(32), Celsius(0).Fahrenheit())
	assert.Equal(t, Celsius(100), Fahrenheit(212).Celsius())
	assert.Equal(t, Fahrenheit(212), Celsius(100).Fahrenheit())
}
