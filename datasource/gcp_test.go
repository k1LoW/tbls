package datasource

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/k1LoW/duration"
)

func TestSetEnvGoogleApplicationCredentials(t *testing.T) {
	tests := []struct {
		values   url.Values
		credsEnv string
		want     string
	}{
		{url.Values{}, "", ""},
		{url.Values{"google_application_credentials": []string{"path/to/creds.json"}}, "", "path/to/creds.json"},
		{url.Values{"credentials": []string{"path/to/creds.json"}}, "", "path/to/creds.json"},
		{url.Values{"creds": []string{"path/to/creds.json"}}, "", "path/to/creds.json"},
		{url.Values{"invalid": []string{"path/to/creds.json"}}, "", ""},
		{url.Values{"creds": []string{"path/to/creds.json"}}, "path/to/creds2.json", "path/to/creds.json"},
		{url.Values{}, "path/to/creds2.json", "path/to/creds2.json"},
	}
	for _, tt := range tests {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")

		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.credsEnv)

		if err := setEnvGoogleApplicationCredentials(tt.values); err != nil {
			t.Fatal(err)
		}

		got := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestGetImpersonateServiceAccountLifetime(t *testing.T) {
	defaultDuration, _ := duration.Parse(defaultImpersonateServiceAccountLifetimeStr)
	specificDuration, _ := duration.Parse("1hour")
	tests := []struct {
		value string
		want  time.Duration
	}{
		{defaultImpersonateServiceAccountLifetimeStr, defaultDuration},
		{"1hour", specificDuration},
	}
	for _, tt := range tests {
		_ = os.Unsetenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT_LIFETIME")
		_ = os.Setenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT_LIFETIME", tt.value)

		got, err := getImpersonateServiceAccountLifetime()
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
