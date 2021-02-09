package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

var testdata = []struct {
	app      int64
	provider int64
	result   int64
}{
	{0, 0, 3600},
	{7200, 0, 7200},
	{0, 7200, 7200},
	{7200, 14400, 7200},
}

func TestSessionDuration(t *testing.T) {
	for _, tc := range testdata {
		viper.Set("apps.test.duration", tc.app)
		viper.Set("apps.test.provider", "test")
		viper.Set("providers.test.duration", tc.provider)

		res := sessionDuration("test")
		if res != tc.result {
			t.Fatalf("Invalid duration: got %v, want: %v", res, tc.result)
		}
	}
}
