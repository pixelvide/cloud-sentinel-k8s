package api

import (
	"strings"
)

// ParseNamespaces parses the namespace query parameter.
// It supports comma-separated values.
// Empty string or "__all__" (in any part) signifies all namespaces, returning []string{""}.
func ParseNamespaces(ns string) []string {
	if ns == "" || ns == "__all__" {
		return []string{""}
	}

	var namespaces []string
	parts := strings.Split(ns, ",")
	for _, n := range parts {
		trimmed := strings.TrimSpace(n)
		if trimmed == "__all__" {
			return []string{""}
		}
		if trimmed != "" {
			namespaces = append(namespaces, trimmed)
		}
	}

	if len(namespaces) == 0 {
		return []string{""}
	}

	return namespaces
}
