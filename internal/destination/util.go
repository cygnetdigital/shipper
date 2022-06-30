package destination

import (
	"fmt"
	"os"
	"strings"
)

func ensureDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0755)
		}

		return err
	}

	return nil
}

// SlugifyServiceName converts a service name to a slug
func SlugifyServiceName(name string) (string, error) {
	name = strings.Replace(name, "service.", "s-", 1)
	name = strings.Replace(name, "api.", "api-", 1)

	if strings.Contains(name, ".") {
		return "", fmt.Errorf("'%s', contains dot after slugification", name)
	}

	return name, nil
}
