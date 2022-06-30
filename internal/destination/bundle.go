package destination

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
)

// DeployContext ...
type DeployContext struct {
	// Name of the project this belongs to
	Project string

	// Name of service. e.g. service.foo
	Name string

	// Slugified name of service. e.g. s-foo
	SlugName string

	// Version of service to deploy
	Version string

	// slugified name of service with version. e.g. s-foo-v1
	SlugNameVersion string

	// Deployment image of service. e.g. gcr.io/foo/service:v1.0.0
	DeployImage string

	// Namespace to put deployment in. e.g. default
	Namespace string
}

// WriteDeployBundle ...
func WriteDeployBundle(templatesDir, manifestRoot string, template string, args *DeployContext) error {
	deployDir := path.Join(manifestRoot, args.Name, args.Version)

	if err := ensureDir(deployDir); err != nil {
		return fmt.Errorf("failed to ensure deploy dir exists '%s': %w", deployDir, err)
	}

	templateDir := path.Join(templatesDir, "deploy", template)

	if err := writeTemplateOut(templateDir, deployDir, args); err != nil {
		return fmt.Errorf("failed to write deploy bundle: %w", err)
	}

	return nil
}

// DeleteDeployBundle ...
func DeleteDeployBundle(manifestRoot string, serviceName, version string) error {
	deployDir := path.Join(manifestRoot, serviceName, version)

	if err := os.RemoveAll(deployDir); err != nil {
		return fmt.Errorf("failed to delete deploy dir '%s': %w", deployDir, err)
	}

	return nil
}

// ReleaseContext ...
type ReleaseContext struct {
	// Name of the project this belongs to
	Project string

	// Name of service. e.g. service.foo
	Name string

	// Slugified name of service. e.g. s-foo
	SlugName string

	// Version of service to release
	Version string

	// slugified name of service with version to release. e.g. s-foo-v1
	SlugNameVersion string

	// Namespace to put service in. e.g. default
	Namespace string
}

var releaseTemplate = `
kind: Service
apiVersion: v1
metadata:
  name: {{ .SlugName }}
  namespace: {{ .Namespace }}
  annotations:
    shipper/bundle: release
    shipper/project: {{ .Project }}
    shipper/service-name: {{ .Name }}
    shipper/current-release: {{ .Version }}
spec:
  ports:
    - name: http
      port: 8000
  selector:
    app: {{ .SlugName }}
    version: {{ .Version }}
`

// WriteReleaseBundle ...
func WriteReleaseBundle(manifestRoot string, args *ReleaseContext) error {
	t, err := template.New("service.yaml").Parse(strings.TrimSpace(releaseTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	releaseDir := path.Join(manifestRoot, args.Name)

	if err := writeTemplate(t, releaseDir, args); err != nil {
		return fmt.Errorf("failed to write release bundle: %w", err)
	}

	return nil
}

// writeTemplateOut to files in the destination directory
func writeTemplateOut(templateDir, destDir string, arg any) error {
	templateFiles := fmt.Sprintf("%s/*.yaml", templateDir)

	t, err := template.New("").ParseGlob(templateFiles)
	if err != nil {
		return fmt.Errorf("failed to parse templates %s: %w", templateFiles, err)
	}

	tpls := t.Templates()

	if len(tpls) == 0 {
		return fmt.Errorf("no templates found %s", templateFiles)
	}

	for _, tp := range tpls {
		if tp.Name() == "" {
			continue
		}

		if err := writeTemplate(tp, destDir, arg); err != nil {
			return fmt.Errorf("failed to write template %s: %w", tp.Name(), err)
		}
	}

	return nil
}

func writeTemplate(tp *template.Template, destDir string, arg any) error {
	f, err := os.Create(path.Join(destDir, tp.Name()))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", tp.Name(), err)
	}

	if err := tp.Execute(f, arg); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", tp.Name(), err)
	}

	return f.Close()
}
