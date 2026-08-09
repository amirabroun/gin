package utils

import (
	"gin/testhelpers"
	"testing"
)

func TestToken(t *testing.T) {
	var id uint = 1

	generatedToken, generateErr := GenerateToken(id)

	parsedId, parseErr := ParseToken(generatedToken)

	testhelpers.AssertNoError(t, generateErr, generatedToken)
	testhelpers.AssertNoError(t, parseErr, parsedId)
}
