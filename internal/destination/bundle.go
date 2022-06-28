package destination

import (
	"fmt"
	"html/template"
	"os"
	"path"
)

// WriteDeployBundle ...
func WriteDeployBundle(templatesDir, manifestRoot string, params *ServiceDeployParams) error {
	templateDir := path.Join(templatesDir, "deploy", params.ServiceConfig.Deploy.Template)
	deployDir := path.Join(manifestRoot, params.Name, fmt.Sprintf("v%d", params.Version))

	if err := ensureDir(deployDir); err != nil {
		return fmt.Errorf("failed to ensure deploy dir exists '%s': %w", deployDir, err)
	}

	if err := writeTemplateOut(templateDir, deployDir, params); err != nil {
		return fmt.Errorf("failed to write deploy bundle: %w", err)
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

		f, err := os.Create(path.Join(destDir, tp.Name()))
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", tp.Name(), err)
		}

		if err := tp.Execute(f, arg); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", tp.Name(), err)
		}
	}

	return nil
}
