package domain

import "errors"

var (
	ErrInvalidInput           = errors.New("invalid input")
	ErrEmailRequired          = errors.New("client email is required")
	ErrReportDeliveryDisabled = errors.New("report delivery is disabled")
	ErrForbidden              = errors.New("forbidden")
	ErrUpstreamResponse       = errors.New("invalid upstream response")
)
