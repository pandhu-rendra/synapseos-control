package docker

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrConnection = errors.New("docker connection failed")
	ErrNotFound   = errors.New("docker resource not found")
	ErrPermission = errors.New("docker permission denied")
)

type Provider interface {
	Ping(ctx context.Context) (PingResult, error)
	Info(ctx context.Context) (HostInfo, error)
	ListContainers(ctx context.Context, state string) ([]Container, error)
	InspectContainer(ctx context.Context, id string) (ContainerDetails, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
	ContainerLogs(ctx context.Context, id string, tail int) ([]LogLine, error)
	Close() error
}

type PingResult struct {
	APIVersion string
	OSType     string
}

type HostInfo struct {
	Name              string `json:"name"`
	ServerVersion     string `json:"server_version"`
	OperatingSystem   string `json:"operating_system"`
	Architecture      string `json:"architecture"`
	CPUs              int    `json:"cpus"`
	MemoryBytes       int64  `json:"memory_bytes"`
	Containers        int    `json:"containers"`
	RunningContainers int    `json:"running_containers"`
	StoppedContainers int    `json:"stopped_containers"`
	Images            int    `json:"images"`
}

type Container struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	State  string `json:"state"`
	Status string `json:"status"`
}

type ContainerDetails struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Image    string          `json:"image"`
	State    string          `json:"state"`
	Status   string          `json:"status"`
	Created  string          `json:"created"`
	Ports    []ContainerPort `json:"ports"`
	Networks []string        `json:"networks"`
}

type ContainerPort struct {
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
	IP          string `json:"ip"`
}

type LogLine struct {
	Stream string `json:"stream"`
	Text   string `json:"text"`
}

func ClassifyError(err error) error {
	if err == nil {
		return nil
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "permission denied") ||
		strings.Contains(message, "operation not permitted") {
		return errors.Join(ErrPermission, err)
	}
	if strings.Contains(message, "no such container") ||
		strings.Contains(message, "not found") {
		return errors.Join(ErrNotFound, err)
	}
	if strings.Contains(message, "connect") ||
		strings.Contains(message, "no such file or directory") ||
		strings.Contains(message, "connection refused") {
		return errors.Join(ErrConnection, err)
	}
	return err
}
