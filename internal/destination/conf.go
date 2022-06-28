package destination

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Service representation in the destination
type Service struct {
	ProjectName string           `yaml:"projectName"`
	ServiceName string           `yaml:"serviceName"`
	Deploys     []*ServiceDeploy `yaml:"deploys"`
}

// NextDeployVersion returns the next deploy version for the given service
func (s *Service) NextDeployVersion() int {
	if len(s.Deploys) == 0 {
		return 1
	}

	return s.Deploys[len(s.Deploys)-1].Version + 1
}

// DeployVersion returns the deploy version for the given service
func (s *Service) DeployVersion() int {
	if len(s.Deploys) == 0 {
		return 0
	}

	return s.Deploys[len(s.Deploys)-1].Version
}

// ServiceDeploy represents a single deploy in the service
type ServiceDeploy struct {
	Version int `yaml:"version"`
}

// ParseStateFiles finds all `shipper.state.yaml` files in the given directory and parse them
func ParseStateFiles(path string) (Services, error) {
	var stateFiles []*Service
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() != "shipper.state.yaml" {
			return nil
		}

		stateFile, err := parseStateFile(path)
		if err != nil {
			return err
		}

		stateFiles = append(stateFiles, stateFile)

		return nil
	})

	return stateFiles, err
}

func parseStateFile(path string) (*Service, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var stateFile Service
	if err := yaml.NewDecoder(file).Decode(&stateFile); err != nil {
		return nil, err
	}

	if err := file.Close(); err != nil {
		return nil, err
	}

	return &stateFile, nil
}

func writeState(state *Service, dir string) error {
	file, err := os.Create(filepath.Join(dir, "shipper.state.yaml"))
	if err != nil {
		return err
	}

	if err := yaml.NewEncoder(file).Encode(state); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

// UpsertServiceStateForDeploy ...
func UpsertServiceStateForDeploy(manifestRoot string, projName string, current *Service, svcParams *ServiceDeployParams) error {
	svcDir := path.Join(manifestRoot, svcParams.Name)
	if err := ensureDir(svcDir); err != nil {
		return fmt.Errorf("failed to ensure deploy dir exists '%s': %w", svcDir, err)
	}

	ds := &ServiceDeploy{
		Version: svcParams.Version,
	}

	if current == nil {
		state := &Service{
			ProjectName: projName,
			ServiceName: svcParams.Name,
			Deploys:     []*ServiceDeploy{ds},
		}

		return writeState(state, svcDir)
	}

	current.Deploys = append(current.Deploys, ds)

	return writeState(current, svcDir)
}
