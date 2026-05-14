package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"omepic/backend/internal/response"
	"omepic/backend/internal/service"
)

type serviceErrorMapping struct {
	Status  int
	Code    string
	Message string
}

func writeServiceError(c *gin.Context, logger *slog.Logger, err error, dependencyLogMessage string, fallbackLogMessage string, overrides map[error]serviceErrorMapping) {
	if mapping, ok := serviceErrorOverride(err, overrides); ok {
		response.Error(c, mapping.Status, mapping.Code, mapping.Message)
		return
	}

	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid_input", service.UserMessage(err, "invalid request"))
	case errors.Is(err, service.ErrConflict):
		response.Error(c, http.StatusConflict, "conflict", service.UserMessage(err, "conflict"))
	case errors.Is(err, service.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not_found", service.UserMessage(err, "resource not found"))
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, service.ErrDependencyUnavailable):
		if logger != nil {
			logger.Error(dependencyLogMessage, "error", err.Error())
		}
		response.Error(c, http.StatusServiceUnavailable, "dependency_unavailable", "dependency unavailable")
	default:
		if logger != nil {
			logger.Error(fallbackLogMessage, "error", err.Error())
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func serviceErrorOverride(err error, overrides map[error]serviceErrorMapping) (serviceErrorMapping, bool) {
	for target, mapping := range overrides {
		if errors.Is(err, target) {
			return mapping, true
		}
	}
	return serviceErrorMapping{}, false
}
