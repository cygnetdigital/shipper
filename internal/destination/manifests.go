package destination

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func openAndParseManifest(filename string) ([]*Manifest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	mfs, err := parseManifest(f)
	if err != nil {
		f.Close()

		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return mfs, nil
}

func parseManifest(f io.Reader) ([]*Manifest, error) {
	out := []*Manifest{}
	decoder := yaml.NewYAMLOrJSONDecoder(f, 4096)

	for {
		ext := Manifest{}

		if err := decoder.Decode(&ext); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		out = append(out, &ext)
	}

	return out, nil
}

func glob(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if filepath.Ext(s) == ".yaml" {
			files = append(files, s)
		}

		return nil
	})

	return files, err
}

// Manifest is an internal type that mostly just passes
// objects through.
type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Raw               []byte
}

// MarshalJSON ...
func (m *Manifest) MarshalJSON() ([]byte, error) {
	return m.Raw, nil
}

// UnmarshalJSON ...
func (m *Manifest) UnmarshalJSON(b []byte) error {
	m.Raw = b

	type Alias Manifest
	aux := Alias{}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	m.TypeMeta = aux.TypeMeta
	m.ObjectMeta = aux.ObjectMeta

	return nil
}

var firstVersionString = "v1"

// NextDeployVersion returns the next deploy version for the given service
func (s *Service) NextDeployVersion() (string, error) {
	if len(s.Deploys) == 0 {
		return firstVersionString, nil
	}

	lastDeploy := s.Deploys[len(s.Deploys)-1]

	i, err := strconv.Atoi(strings.TrimPrefix(lastDeploy.Version, "v"))
	if err != nil {
		return "", fmt.Errorf("failed to parse last deploy version '%s': %w", lastDeploy.Version, err)
	}

	return fmt.Sprintf("v%d", i+1), nil
}
