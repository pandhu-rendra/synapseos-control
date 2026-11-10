package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/synapseos/synapseos-control/apps/api/internal/docker"
)

func TestHealthEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload responseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error != nil {
		t.Fatalf("expected no error, got %+v", payload.Error)
	}
}

func TestAPITokenMiddleware(t *testing.T) {
	handler := withAPIToken(newRouter(&fakeDockerProvider{}), "test-token")

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/docker/status", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, unauthorized.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/docker/status", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer test-token")
	authorized := httptest.NewRecorder()
	handler.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, authorized.Code)
	}
}

func TestHealthEndpointDoesNotRequireAPIToken(t *testing.T) {
	handler := withAPIToken(newRouter(&fakeDockerProvider{}), "test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestOptionsRequestDoesNotRequireAPIToken(t *testing.T) {
	handler := withAPIToken(newRouter(&fakeDockerProvider{}), "test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodOptions, "/api/v1/docker/status", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestDockerStatusEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/docker/status", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{
		pingResult: docker.PingResult{APIVersion: "1.45", OSType: "linux"},
	}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestDockerStatusPermissionError(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/docker/status", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{
		pingError: docker.ErrPermission,
	}).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}

	var payload responseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error == nil || payload.Error.Code != "DOCKER_PERMISSION_DENIED" {
		t.Fatalf("expected permission error, got %+v", payload.Error)
	}
}

func TestContainersEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers?state=running", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{
		containers: []docker.Container{{ID: "abc123", Name: "web", Image: "nginx:latest", State: "running", Status: "Up 2 minutes"}},
	}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	var payload responseEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := payload.Data.([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("expected one container, got %#v", payload.Data)
	}
}

func TestContainersRejectInvalidState(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers?state=unknown", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestContainerDetailsEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers/abc123", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{
		details: docker.ContainerDetails{
			ID: "abc123", Name: "web", Image: "nginx:latest",
			State: "running", Created: "2026-11-10T00:00:00Z",
			Networks: []string{"bridge"},
		},
	}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestContainerDetailsRejectInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers/%2Fetc%2Fpasswd", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestContainerDetailsNotFound(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers/missing", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{detailsError: docker.ErrNotFound}).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestContainerActionEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/containers/abc123/restart", nil)
	response := httptest.NewRecorder()
	provider := &fakeDockerProvider{}

	newRouter(provider).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if provider.action != "restart" {
		t.Fatalf("expected restart action, got %q", provider.action)
	}
}

func TestContainerActionRejectsUnsupportedAction(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/containers/abc123/delete", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{}).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestContainerLogsEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers/abc123/logs?tail=50", nil)
	response := httptest.NewRecorder()
	provider := &fakeDockerProvider{
		logs: []docker.LogLine{{Stream: "stdout", Text: "ready"}},
	}

	newRouter(provider).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if provider.logsTail != 50 {
		t.Fatalf("expected tail 50, got %d", provider.logsTail)
	}
}

func TestContainerLogsRejectInvalidTail(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/containers/abc123/logs?tail=1001", nil)
	response := httptest.NewRecorder()

	newRouter(&fakeDockerProvider{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestDockerSettingsRejectRemoteHost(t *testing.T) {
	settings := docker.Settings{Host: "tcp://127.0.0.1:2375", Timeout: 10}
	if err := docker.ValidateSettings(settings); err == nil {
		t.Fatal("expected remote host to be rejected")
	}
}

func TestDockerSettingsEndpoints(t *testing.T) {
	manager, err := docker.NewManager(docker.Settings{
		Host:    "unix:///tmp/synapseos-control-http.sock",
		Timeout: 10,
	})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	defer manager.Close()

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/settings/docker", nil)
	getResponse := httptest.NewRecorder()
	newRouter(manager).ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected GET status %d, got %d", http.StatusOK, getResponse.Code)
	}

	putRequest := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/settings/docker",
		strings.NewReader(`{"host":"unix:///tmp/updated.sock","timeout_seconds":5}`),
	)
	putResponse := httptest.NewRecorder()
	newRouter(manager).ServeHTTP(putResponse, putRequest)
	if putResponse.Code != http.StatusOK {
		t.Fatalf("expected PUT status %d, got %d", http.StatusOK, putResponse.Code)
	}
	if got := manager.Settings(); got.Host != "unix:///tmp/updated.sock" || got.Timeout != 5 {
		t.Fatalf("unexpected settings after PUT: %+v", got)
	}
}

func TestDockerSettingsRejectMalformedPayload(t *testing.T) {
	manager, err := docker.NewManager(docker.Settings{
		Host:    "unix:///tmp/synapseos-control-http.sock",
		Timeout: 10,
	})
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	defer manager.Close()

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/settings/docker",
		strings.NewReader(`{"host":`),
	)
	response := httptest.NewRecorder()
	newRouter(manager).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

type fakeDockerProvider struct {
	pingResult   docker.PingResult
	pingError    error
	infoResult   docker.HostInfo
	infoError    error
	containers   []docker.Container
	listError    error
	details      docker.ContainerDetails
	detailsError error
	actionError  error
	action       string
	logs         []docker.LogLine
	logsTail     int
	logsError    error
}

func (provider *fakeDockerProvider) Ping(context.Context) (docker.PingResult, error) {
	return provider.pingResult, provider.pingError
}

func (provider *fakeDockerProvider) Info(context.Context) (docker.HostInfo, error) {
	return provider.infoResult, provider.infoError
}

func (provider *fakeDockerProvider) ListContainers(context.Context, string) ([]docker.Container, error) {
	return provider.containers, provider.listError
}

func (provider *fakeDockerProvider) InspectContainer(context.Context, string) (docker.ContainerDetails, error) {
	return provider.details, provider.detailsError
}

func (provider *fakeDockerProvider) StartContainer(context.Context, string) error {
	provider.action = "start"
	return provider.actionError
}

func (provider *fakeDockerProvider) StopContainer(context.Context, string) error {
	provider.action = "stop"
	return provider.actionError
}

func (provider *fakeDockerProvider) RestartContainer(context.Context, string) error {
	provider.action = "restart"
	return provider.actionError
}

func (provider *fakeDockerProvider) ContainerLogs(_ context.Context, _ string, tail int) ([]docker.LogLine, error) {
	provider.logsTail = tail
	return provider.logs, provider.logsError
}

func (provider *fakeDockerProvider) Close() error {
	return nil
}

var _ docker.Provider = (*fakeDockerProvider)(nil)
