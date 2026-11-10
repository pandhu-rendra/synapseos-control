package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/synapseos/synapseos-control/apps/api/internal/docker"
)

type responseEnvelope struct {
	Data  any       `json:"data"`
	Error *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type dockerStatus struct {
	Connected  bool   `json:"connected"`
	APIVersion string `json:"api_version,omitempty"`
	OSType     string `json:"os_type,omitempty"`
}

var validContainerStates = map[string]bool{
	"created":    true,
	"running":    true,
	"paused":     true,
	"restarting": true,
	"removing":   true,
	"exited":     true,
	"dead":       true,
}

var containerIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

func newRouter(provider docker.Provider) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handleHealth)
	mux.HandleFunc("GET /api/v1/docker/status", func(w http.ResponseWriter, r *http.Request) {
		handleDockerStatus(w, r, provider)
	})
	mux.HandleFunc("GET /api/v1/docker/info", func(w http.ResponseWriter, r *http.Request) {
		handleDockerInfo(w, r, provider)
	})
	if manager, ok := provider.(*docker.Manager); ok {
		mux.HandleFunc("GET /api/v1/settings/docker", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, responseEnvelope{Data: manager.Settings()})
		})
		mux.HandleFunc("PUT /api/v1/settings/docker", func(w http.ResponseWriter, r *http.Request) {
			handleDockerSettings(w, r, manager)
		})
	}
	mux.HandleFunc("GET /api/v1/containers", func(w http.ResponseWriter, r *http.Request) {
		handleContainers(w, r, provider)
	})
	mux.HandleFunc("GET /api/v1/containers/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleContainerDetails(w, r, provider)
	})
	mux.HandleFunc("POST /api/v1/containers/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		handleContainerAction(w, r, provider)
	})
	mux.HandleFunc("GET /api/v1/containers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
		handleContainerLogs(w, r, provider)
	})
	return withCORS(mux)
}

func handleDockerSettings(w http.ResponseWriter, request *http.Request, manager *docker.Manager) {
	var settings docker.Settings
	decoder := json.NewDecoder(http.MaxBytesReader(w, request.Body, 4096))
	if err := decoder.Decode(&settings); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{Code: "INVALID_SETTINGS", Message: "The Docker settings payload is invalid"},
		})
		return
	}
	if err := docker.ValidateSettings(settings); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{Code: "INVALID_SETTINGS", Message: err.Error()},
		})
		return
	}
	if err := manager.UpdateSettings(settings); err != nil {
		writeDockerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{Data: manager.Settings()})
}

func handleContainerLogs(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	id := request.PathValue("id")
	if !containerIDPattern.MatchString(id) {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{Code: "INVALID_CONTAINER_ID", Message: "The container identifier is invalid"},
		})
		return
	}

	tail := 200
	if rawTail := strings.TrimSpace(request.URL.Query().Get("tail")); rawTail != "" {
		parsedTail, err := strconv.Atoi(rawTail)
		if err != nil || parsedTail < 1 || parsedTail > 1000 {
			writeJSON(w, http.StatusBadRequest, responseEnvelope{
				Error: &apiError{Code: "INVALID_LOG_TAIL", Message: "The log tail must be a number from 1 to 1000"},
			})
			return
		}
		tail = parsedTail
	}

	logs, err := provider.ContainerLogs(request.Context(), id, tail)
	if err != nil {
		if errors.Is(err, docker.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, responseEnvelope{
				Error: &apiError{Code: "CONTAINER_NOT_FOUND", Message: "The requested container was not found"},
			})
			return
		}
		writeDockerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{Data: logs})
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{
		Data: map[string]string{"status": "ok"},
	})
}

func handleDockerStatus(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	result, err := provider.Ping(request.Context())
	if err != nil {
		writeDockerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{
		Data: dockerStatus{
			Connected:  true,
			APIVersion: result.APIVersion,
			OSType:     result.OSType,
		},
	})
}

func handleDockerInfo(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	info, err := provider.Info(request.Context())
	if err != nil {
		writeDockerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: info})
}

func handleContainers(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	state := strings.TrimSpace(request.URL.Query().Get("state"))
	if state != "" && !validContainerStates[state] {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{
				Code:    "INVALID_CONTAINER_STATE",
				Message: "The container state filter is not supported",
			},
		})
		return
	}

	containers, err := provider.ListContainers(request.Context(), state)
	if err != nil {
		writeDockerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: containers})
}

func handleContainerDetails(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	id := request.PathValue("id")
	if !containerIDPattern.MatchString(id) {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{
				Code:    "INVALID_CONTAINER_ID",
				Message: "The container identifier is invalid",
			},
		})
		return
	}

	details, err := provider.InspectContainer(request.Context(), id)
	if err != nil {
		if errors.Is(err, docker.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, responseEnvelope{
				Error: &apiError{
					Code:    "CONTAINER_NOT_FOUND",
					Message: "The requested container was not found",
				},
			})
			return
		}
		writeDockerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{Data: details})
}

func handleContainerAction(w http.ResponseWriter, request *http.Request, provider docker.Provider) {
	id := request.PathValue("id")
	action := request.PathValue("action")
	if !containerIDPattern.MatchString(id) {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{
			Error: &apiError{Code: "INVALID_CONTAINER_ID", Message: "The container identifier is invalid"},
		})
		return
	}

	var err error
	switch action {
	case "start":
		err = provider.StartContainer(request.Context(), id)
	case "stop":
		err = provider.StopContainer(request.Context(), id)
	case "restart":
		err = provider.RestartContainer(request.Context(), id)
	default:
		writeJSON(w, http.StatusNotFound, responseEnvelope{
			Error: &apiError{Code: "UNSUPPORTED_CONTAINER_ACTION", Message: "The container action is not supported"},
		})
		return
	}
	if err != nil {
		if errors.Is(err, docker.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, responseEnvelope{
				Error: &apiError{Code: "CONTAINER_NOT_FOUND", Message: "The requested container was not found"},
			})
			return
		}
		writeDockerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, responseEnvelope{
		Data: map[string]string{"id": id, "action": action, "status": "completed"},
	})
}

func writeDockerError(w http.ResponseWriter, err error) {
	status := http.StatusServiceUnavailable
	code := "DOCKER_CONNECTION_FAILED"
	message := "Unable to connect to Docker Engine"
	if errors.Is(err, docker.ErrPermission) {
		code = "DOCKER_PERMISSION_DENIED"
		message = "The application cannot access the Docker socket"
	}
	writeJSON(w, status, responseEnvelope{
		Error: &apiError{Code: code, Message: message},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload responseEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withAPIToken(next http.Handler, expectedToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}
		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		if len(authorization) <= len(prefix) ||
			!strings.EqualFold(authorization[:len(prefix)], prefix) ||
			subtle.ConstantTimeCompare([]byte(authorization[len(prefix):]), []byte(expectedToken)) != 1 {
			writeJSON(w, http.StatusUnauthorized, responseEnvelope{
				Error: &apiError{Code: "UNAUTHORIZED", Message: "A valid API token is required"},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
