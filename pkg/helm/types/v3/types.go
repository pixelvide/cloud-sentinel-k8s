package v3

import "time"

type HelmRelease struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Revision   int       `json:"revision"`
	Status     string    `json:"status"`
	Chart      string    `json:"chart"`
	AppVersion string    `json:"app_version"`
	Updated    time.Time `json:"updated"`
	Values     string    `json:"values,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	Manifest   string    `json:"manifest,omitempty"`
}
