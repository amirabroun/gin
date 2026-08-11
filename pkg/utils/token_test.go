package utils

import "testing"

func TestToken(t *testing.T) {
	var id uint = 1

	generatedToken, generateErr := GenerateToken(id)

	if generateErr != nil {
		t.Fatalf("unexpected error: %v", generateErr)
	}

	t.Log(generatedToken)

	parsedId, parseErr := ParseToken(generatedToken)

	if parseErr != nil {
		t.Fatalf("unexpected error: %v", generateErr)
	}

	if id != parsedId {
		t.Fatal("token is not created in correct")
	}

	t.Log(parsedId)
}
