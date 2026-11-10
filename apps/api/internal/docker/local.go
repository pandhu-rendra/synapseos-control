package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

type LocalProvider struct {
	client  *client.Client
	timeout time.Duration
}

func NewLocalProvider(host string, timeout time.Duration) (*LocalProvider, error) {
	apiClient, err := client.New(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
		client.WithTimeout(timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &LocalProvider{client: apiClient, timeout: timeout}, nil
}

func (provider *LocalProvider) Ping(ctx context.Context) (PingResult, error) {
	pingContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	result, err := provider.client.Ping(pingContext, client.PingOptions{
		NegotiateAPIVersion: true,
	})
	if err != nil {
		return PingResult{}, ClassifyError(fmt.Errorf("ping docker: %w", err))
	}

	return PingResult{
		APIVersion: result.APIVersion,
		OSType:     result.OSType,
	}, nil
}

func (provider *LocalProvider) Info(ctx context.Context) (HostInfo, error) {
	infoContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	result, err := provider.client.Info(infoContext, client.InfoOptions{})
	if err != nil {
		return HostInfo{}, ClassifyError(fmt.Errorf("retrieve docker info: %w", err))
	}

	return HostInfo{
		Name:              result.Info.Name,
		ServerVersion:     result.Info.ServerVersion,
		OperatingSystem:   result.Info.OperatingSystem,
		Architecture:      result.Info.Architecture,
		CPUs:              result.Info.NCPU,
		MemoryBytes:       result.Info.MemTotal,
		Containers:        result.Info.Containers,
		RunningContainers: result.Info.ContainersRunning,
		StoppedContainers: result.Info.ContainersStopped,
		Images:            result.Info.Images,
	}, nil
}

func (provider *LocalProvider) ListContainers(ctx context.Context, state string) ([]Container, error) {
	listContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	options := client.ContainerListOptions{All: true}
	if state != "" {
		options.Filters = client.Filters{}.Add("status", state)
	}

	result, err := provider.client.ContainerList(listContext, options)
	if err != nil {
		return nil, ClassifyError(fmt.Errorf("list docker containers: %w", err))
	}

	containers := make([]Container, 0, len(result.Items))
	for _, item := range result.Items {
		name := ""
		if len(item.Names) > 0 {
			name = item.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}
		containers = append(containers, Container{
			ID:     item.ID,
			Name:   name,
			Image:  item.Image,
			State:  string(item.State),
			Status: item.Status,
		})
	}

	return containers, nil
}

func (provider *LocalProvider) InspectContainer(ctx context.Context, id string) (ContainerDetails, error) {
	inspectContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	result, err := provider.client.ContainerInspect(inspectContext, id, client.ContainerInspectOptions{})
	if err != nil {
		return ContainerDetails{}, ClassifyError(fmt.Errorf("inspect docker container: %w", err))
	}

	container := result.Container
	details := ContainerDetails{
		ID:       container.ID,
		Name:     container.Name,
		Image:    container.Config.Image,
		Created:  container.Created,
		Ports:    make([]ContainerPort, 0),
		Networks: make([]string, 0),
	}
	if details.Name != "" && details.Name[0] == '/' {
		details.Name = details.Name[1:]
	}
	if container.State != nil {
		details.State = string(container.State.Status)
		details.Status = container.State.Error
		if details.Status == "" {
			details.Status = details.State
		}
	}
	if container.NetworkSettings != nil {
		for name := range container.NetworkSettings.Networks {
			details.Networks = append(details.Networks, name)
		}
		sort.Strings(details.Networks)
		for port, bindings := range container.NetworkSettings.Ports {
			for _, binding := range bindings {
				publicPort, _ := strconv.ParseUint(binding.HostPort, 10, 16)
				details.Ports = append(details.Ports, ContainerPort{
					PrivatePort: port.Num(),
					PublicPort:  uint16(publicPort),
					Type:        string(port.Proto()),
					IP:          binding.HostIP.String(),
				})
			}
			if len(bindings) == 0 {
				details.Ports = append(details.Ports, ContainerPort{
					PrivatePort: port.Num(),
					Type:        string(port.Proto()),
				})
			}
		}
	}

	return details, nil
}

func (provider *LocalProvider) StartContainer(ctx context.Context, id string) error {
	actionContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()
	_, err := provider.client.ContainerStart(actionContext, id, client.ContainerStartOptions{})
	if err != nil {
		return ClassifyError(fmt.Errorf("start docker container: %w", err))
	}
	return nil
}

func (provider *LocalProvider) StopContainer(ctx context.Context, id string) error {
	actionContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()
	_, err := provider.client.ContainerStop(actionContext, id, client.ContainerStopOptions{})
	if err != nil {
		return ClassifyError(fmt.Errorf("stop docker container: %w", err))
	}
	return nil
}

func (provider *LocalProvider) RestartContainer(ctx context.Context, id string) error {
	actionContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()
	_, err := provider.client.ContainerRestart(actionContext, id, client.ContainerRestartOptions{})
	if err != nil {
		return ClassifyError(fmt.Errorf("restart docker container: %w", err))
	}
	return nil
}

func (provider *LocalProvider) ContainerLogs(ctx context.Context, id string, tail int) ([]LogLine, error) {
	logContext, cancel := context.WithTimeout(ctx, provider.timeout)
	defer cancel()

	result, err := provider.client.ContainerLogs(logContext, id, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       fmt.Sprintf("%d", tail),
	})
	if err != nil {
		return nil, ClassifyError(fmt.Errorf("retrieve docker container logs: %w", err))
	}
	defer result.Close()

	var stdout, stderr io.Writer = &strings.Builder{}, &strings.Builder{}
	if _, err := stdcopy.StdCopy(stdout, stderr, result); err != nil {
		return nil, ClassifyError(fmt.Errorf("decode docker container logs: %w", err))
	}
	lines := make([]LogLine, 0, tail)
	appendLines := func(stream, output string) {
		scanner := bufio.NewScanner(strings.NewReader(output))
		for scanner.Scan() {
			lines = append(lines, LogLine{Stream: stream, Text: scanner.Text()})
		}
	}
	appendLines("stdout", stdout.(*strings.Builder).String())
	appendLines("stderr", stderr.(*strings.Builder).String())
	if len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return lines, nil
}

func (provider *LocalProvider) Close() error {
	return provider.client.Close()
}

var _ Provider = (*LocalProvider)(nil)
var _ io.Closer = (*LocalProvider)(nil)
