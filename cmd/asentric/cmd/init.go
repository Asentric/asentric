package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed all:templates
var templateFS embed.FS

// TemplateData holds data for template rendering.
type TemplateData struct {
	ProjectName string
	ModulePath  string
}

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Create a new Asentric project",
	Long: `Create a new Asentric project with the recommended structure.

This command generates a complete project scaffold including:
  • Configuration files (asentric.yaml, registry.yaml)
  • Example detection rule
  • Main entry point with runtime wiring
  • README with documentation

Example:
  asentric init my-watcher
  cd my-watcher
  go mod tidy
  go run cmd/watcher/main.go`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(projectName, " /\\:*?\"<>|") {
		return fmt.Errorf("project name contains invalid characters")
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	fmt.Printf("Creating new Asentric project: %s\n\n", projectName)

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		ProjectName: projectName,
		ModulePath:  projectName,
	}

	// Copy template files
	if err := copyTemplates(projectName, data); err != nil {
		// Cleanup on error
		os.RemoveAll(projectName)
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Print success message
	fmt.Println()
	fmt.Println("Project created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  # Edit config/asentric.yaml with your settings")
	fmt.Println("  go run cmd/watcher/main.go")
	fmt.Println()

	return nil
}

func copyTemplates(projectDir string, data TemplateData) error {
	// Define file mappings: source template -> destination
	files := map[string]string{
		"templates/config/asentric.yaml":       "config/asentric.yaml",
		"templates/config/registry.yaml":       "config/registry.yaml",
		"templates/rules/example_rule.go.tmpl": "rules/example_rule.go",
		"templates/cmd/watcher/main.go.tmpl":   "cmd/watcher/main.go",
		"templates/go.mod.tmpl":                "go.mod",
		"templates/README.md.tmpl":             "README.md",
	}

	for src, dst := range files {
		// Read template file
		content, err := templateFS.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", src, err)
		}

		// Create destination directory
		dstPath := filepath.Join(projectDir, dst)
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
		}

		// Process template if it's a .tmpl file or needs variable substitution
		var output []byte
		if strings.HasSuffix(src, ".tmpl") || strings.HasSuffix(dst, ".md") || strings.HasSuffix(dst, "go.mod") {
			tmpl, err := template.New(filepath.Base(src)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", src, err)
			}

			var buf strings.Builder
			if err := tmpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("failed to execute template %s: %w", src, err)
			}
			output = []byte(buf.String())
		} else {
			output = content
		}

		// Write file
		if err := os.WriteFile(dstPath, output, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", dstPath, err)
		}

		fmt.Printf("Created: %s\n", dst)
	}

	// Create empty directories with .gitkeep
	emptyDirs := []string{"abi"}
	for _, dir := range emptyDirs {
		dirPath := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		// Create .gitkeep
		gitkeep := filepath.Join(dirPath, ".gitkeep")
		if err := os.WriteFile(gitkeep, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create .gitkeep: %w", err)
		}
		fmt.Printf("Created: %s/\n", dir)
	}

	return nil
}
