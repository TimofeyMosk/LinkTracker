package httpapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"LinkTracker/internal/infrastructure/httpapi"
)

func Test_GetTgIDFromString(t *testing.T) {
	testcases := []struct {
		testName    string
		input       string
		expectedID  int64
		expectedErr bool
	}{
		{testName: "validTgID", input: "123", expectedID: 123, expectedErr: false},
		{testName: "NotValidTgID", input: "qwery", expectedID: 0, expectedErr: true},
		{testName: "EmptyString", input: "", expectedID: 0, expectedErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			result, err := httpapi.GetTgIDFromString(tc.input)
			assert.Equal(t, tc.expectedErr, err != nil)
			assert.Equal(t, tc.expectedID, result)
		})
	}
}
