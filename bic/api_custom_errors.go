package bic

import (
	"net/http"

	"github.com/mercadolibre/go-meli-toolkit/goutils/apierrors"
)

func NewNotAcceptableApiError(message string, err error) apierrors.ApiError {
	cause := apierrors.CauseList{}
	if err != nil {
		cause = append(cause, err.Error())
	}
	return apierrors.NewApiError(message, "not_acceptable", http.StatusNotAcceptable, cause)
}

func NewLockedApiError(message string, err error) apierrors.ApiError {
	cause := apierrors.CauseList{}
	if err != nil {
		cause = append(cause, err.Error())
	}
	return apierrors.NewApiError(message, "locked", http.StatusLocked, cause)
}

func NewPostToReprocessingQueueFail(message string, err error) apierrors.ApiError {
	cause := apierrors.CauseList{}
	if err != nil {
		cause = append(cause, err.Error())
	}
	return apierrors.NewApiError(message, "failure while posting to reprocessing queue", http.StatusInternalServerError, cause)
}
