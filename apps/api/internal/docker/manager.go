package docker

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"
)

type Settings struct {
	Host    string `json:"host"`
	Timeout int    `json:"timeout_seconds"`
}

type Manager struct {
	mu       sync.RWMutex
	provider *LocalProvider
	settings Settings
}

func NewManager(settings Settings) (*Manager, error) {
	provider, err := newProvider(settings)
	if err != nil {
		return nil, err
	}
	return &Manager{provider: provider, settings: settings}, nil
}

func (manager *Manager) Settings() Settings {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.settings
}

func (manager *Manager) UpdateSettings(settings Settings) error {
	if err := ValidateSettings(settings); err != nil {
		return err
	}
	provider, err := newProvider(settings)
	if err != nil {
		return err
	}

	manager.mu.Lock()
	oldProvider := manager.provider
	manager.provider = provider
	manager.settings = settings
	manager.mu.Unlock()
	return oldProvider.Close()
}

func ValidateSettings(settings Settings) error {
	if settings.Timeout < 1 || settings.Timeout > 120 {
		return errors.New("timeout must be between 1 and 120 seconds")
	}
	parsed, err := url.Parse(settings.Host)
	if err != nil || parsed.Scheme != "unix" || parsed.Path == "" {
		return errors.New("host must be a local unix socket endpoint")
	}
	return nil
}

func newProvider(settings Settings) (*LocalProvider, error) {
	return NewLocalProvider(settings.Host, time.Duration(settings.Timeout)*time.Second)
}

func (manager *Manager) withProvider(fn func(Provider) error) error {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return fn(manager.provider)
}

func (manager *Manager) Ping(ctx context.Context) (PingResult, error) {
	var result PingResult
	err := manager.withProvider(func(provider Provider) error {
		var err error
		result, err = provider.Ping(ctx)
		return err
	})
	return result, err
}

func (manager *Manager) Info(ctx context.Context) (HostInfo, error) {
	var result HostInfo
	err := manager.withProvider(func(provider Provider) error {
		var err error
		result, err = provider.Info(ctx)
		return err
	})
	return result, err
}

func (manager *Manager) ListContainers(ctx context.Context, state string) ([]Container, error) {
	var result []Container
	err := manager.withProvider(func(provider Provider) error {
		var err error
		result, err = provider.ListContainers(ctx, state)
		return err
	})
	return result, err
}

func (manager *Manager) InspectContainer(ctx context.Context, id string) (ContainerDetails, error) {
	var result ContainerDetails
	err := manager.withProvider(func(provider Provider) error {
		var err error
		result, err = provider.InspectContainer(ctx, id)
		return err
	})
	return result, err
}

func (manager *Manager) StartContainer(ctx context.Context, id string) error {
	return manager.withProvider(func(provider Provider) error {
		return provider.StartContainer(ctx, id)
	})
}

func (manager *Manager) StopContainer(ctx context.Context, id string) error {
	return manager.withProvider(func(provider Provider) error {
		return provider.StopContainer(ctx, id)
	})
}

func (manager *Manager) RestartContainer(ctx context.Context, id string) error {
	return manager.withProvider(func(provider Provider) error {
		return provider.RestartContainer(ctx, id)
	})
}

func (manager *Manager) ContainerLogs(ctx context.Context, id string, tail int) ([]LogLine, error) {
	var result []LogLine
	err := manager.withProvider(func(provider Provider) error {
		var err error
		result, err = provider.ContainerLogs(ctx, id, tail)
		return err
	})
	return result, err
}

func (manager *Manager) Close() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.provider.Close()
}

var _ Provider = (*Manager)(nil)
