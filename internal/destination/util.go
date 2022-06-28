package destination

import "os"

func ensureDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0755)
		}

		return err
	}

	return nil
}
