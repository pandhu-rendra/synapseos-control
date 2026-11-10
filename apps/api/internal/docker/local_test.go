package docker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "permission",
			err:      errors.New("dial unix /var/run/docker.sock: permission denied"),
			expected: ErrPermission,
		},
		{
			name:     "connection",
			err:      errors.New("dial unix /var/run/docker.sock: connect: no such file or directory"),
			expected: ErrConnection,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classified := ClassifyError(test.err)
			if !errors.Is(classified, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, classified)
			}
		})
	}
}

func TestLocalProviderUnavailable(t *testing.T) {
	provider, err := NewLocalProvider("unix:///tmp/synapseos-control-missing.sock", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	defer provider.Close()

	_, err = provider.Ping(context.Background())
	if err == nil {
		t.Fatal("expected ping to fail")
	}
	if !errors.Is(err, ErrConnection) {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestManagerSettingsAndUpdate(t *testing.T) {
	manager, err := NewManager(Settings{
		Host:    "unix:///tmp/synapseos-control-manager.sock",
		Timeout: 2,
	})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	defer manager.Close()

	if got := manager.Settings(); got.Host != "unix:///tmp/synapseos-control-manager.sock" || got.Timeout != 2 {
		t.Fatalf("unexpected initial settings: %+v", got)
	}

	updated := Settings{
		Host:    "unix:///tmp/synapseos-control-manager-updated.sock",
		Timeout: 5,
	}
	if err := manager.UpdateSettings(updated); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if got := manager.Settings(); got != updated {
		t.Fatalf("unexpected updated settings: %+v", got)
	}
}

func TestValidateSettingsAcceptsUnixSocket(t *testing.T) {
	if err := ValidateSettings(Settings{
		Host:    "unix:///var/run/docker.sock",
		Timeout: 10,
	}); err != nil {
		t.Fatalf("expected unix socket to be accepted: %v", err)
	}
}
