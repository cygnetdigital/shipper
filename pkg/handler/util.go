package handler

import (
	"fmt"
	"strings"
)

// SlugifyServiceName converts a service name to a slug
func SlugifyServiceName(name string) (string, error) {
	name = strings.Replace(name, "service.", "s-", 1)
	name = strings.Replace(name, "api.", "api-", 1)

	if strings.Contains(name, ".") {
		return "", fmt.Errorf("'%s', contains dot after slugification", name)
	}

	return name, nil
}
