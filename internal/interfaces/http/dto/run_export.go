package dto

import "time"

type RequestRunExportRequest struct {
	From *time.Time `json:"from" validate:"omitempty"`
	To   *time.Time `json:"to" validate:"omitempty"`
}
