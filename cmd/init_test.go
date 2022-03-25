package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func getProject() *Project {
	wd, _ := os.Getwd()
	return &Project{
		AbsolutePath: fmt.Sprintf("%s/testproject", wd),
		Legal:        getLicense(),
		Copyright:    copyrightLine(),
		AppName:      "cmd",
		PkgName:      "github.com/spf13/cobra-cli/cmd/cmd",
		Viper:        true,
	}
}

func TestGoldenInitCmd(t *testing.T) {

	dir, err := ioutil.TempDir("", "cobra-init")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tests := []struct {
		name      string
		args      []string
		pkgName   string
		expectErr bool
	}{
		{
			name:      "successfully creates a project based on module",
			args:      []string{"testproject"},
			pkgName:   "github.com/spf13/testproject",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			viper.Set("useViper", true)
			viper.Set("license", "apache")
			projectPath, err := initializeProject(tt.args)
			defer func() {
				if projectPath != "" {
					os.RemoveAll(projectPath)
				}
			}()

			if !tt.expectErr && err != nil {
				t.Fatalf("did not expect an error, got %s", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error but got none")
				} else {
					// got an expected error nothing more to do
					return
				}
			}

			expectedFiles := []string{"LICENSE", "main.go", "cmd/root.go"}
			for _, f := range expectedFiles {
				generatedFile := fmt.Sprintf("%s/%s", projectPath, f)
				goldenFile := fmt.Sprintf("testdata/%s.golden", filepath.Base(f))
				err := compareFiles(generatedFile, goldenFile)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestGoldenInitCmdWithinWorkspace(t *testing.T) {

	dir, err := ioutil.TempDir("", "cobra-init")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cmd := exec.Command("go", "work", "init")
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		pkgName   string
		expectErr bool
	}{
		{
			name:      "successfully initializes an existing project within a workspace",
			args:      []string{"testproject"},
			pkgName:   "github.com/spf13/testproject",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preExistingModule := "pre-existing-fake-project"
			makeModuleInWorkspace(t, dir, preExistingModule)
			makeModuleInWorkspace(t, dir, tt.args[0])
			err := os.Chdir(dir)
			if err != nil {
				t.Fatal(err)
			}

			viper.Set("useViper", true)
			viper.Set("license", "apache")
			projectPath, err := initializeProject(tt.args)
			defer func() {
				if projectPath != "" {
					os.RemoveAll(projectPath)
				}
			}()

			if !tt.expectErr && err != nil {
				t.Fatalf("did not expect an error, got %s", err)
			}
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error but got none")
				} else {
					// got an expected error nothing more to do
					return
				}
			}

			expectedFiles := []string{"LICENSE", "main.go", "cmd/root.go"}
			for _, f := range expectedFiles {
				generatedFile := fmt.Sprintf("%s/%s", projectPath, f)
				goldenFile := fmt.Sprintf("testdata/%s.golden", filepath.Base(f))
				err := compareFiles(generatedFile, goldenFile)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func makeModuleInWorkspace(t *testing.T, workspaceDir, moduleName string) {
	cmd := exec.Command("mkdir", moduleName)
	cmd.Dir = workspaceDir
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = path.Join(workspaceDir, moduleName)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("go", "work", "use", moduleName)
	cmd.Dir = workspaceDir
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
