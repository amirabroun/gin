package testhelpers

import "testing"

func AssertNoError(t *testing.T, err error, data any) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	t.Log(data)
}
