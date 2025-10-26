package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func isInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	dir = filepath.Clean(dir)

	for _, p := range paths {
		cleanPath := filepath.Clean(p)
		if cleanPath == dir {
			return true
		}
	}
	return false
}

func Build() error {
	fmt.Println("Building zzk...")

	fmt.Println("Running go vet...")
	if err := sh.Run("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	if err := os.MkdirAll("bin", 0755); err != nil {
		return err
	}
	binary := "bin/zzk"
	if runtime.GOOS == "windows" {
		binary = "bin/zzk.exe"
	}
	return sh.Run("go", "build", "-o", binary, ".")
}

func getInstallDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	bioDir := homeDir + "/.bio/bin"
	if info, err := os.Stat(bioDir); err == nil && info.IsDir() {
		return bioDir, nil
	}

	var candidateDir string
	switch runtime.GOOS {
	case "linux":
		candidateDir = homeDir + "/.local/bin"
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = homeDir + "\\AppData\\Local"
		}
		candidateDir = localAppData + "\\Microsoft\\WindowsApps"
	case "darwin":
		return "", fmt.Errorf("on macOS, please create ~/.bio/bin first, or use sudo to install to /usr/local/bin")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Only check PATH for platform-specific defaults
	if !isInPath(candidateDir) {
		return "", fmt.Errorf("installation directory %s is not in PATH - please create ~/.bio/bin and add it to your PATH, or add %s to your PATH", candidateDir, candidateDir)
	}

	return candidateDir, nil
}

func Install() error {
	fmt.Println("Installing zzk...")

	installDir, err := getInstallDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	mg.Deps(Build)

	binary := "zzk"
	if runtime.GOOS == "windows" {
		binary = "zzk.exe"
	}

	src := "bin/" + binary
	dst := installDir + "/" + binary
	if runtime.GOOS == "windows" {
		dst = installDir + "\\" + binary
	}

	if err := sh.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(dst, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}

	fmt.Printf("✓ Installed to %s\n", dst)
	return nil
}

func Clean() error {
	fmt.Println("Cleaning...")
	return sh.Rm("bin")
}

func Vet() error {
	fmt.Println("Running go vet...")
	return sh.Run("go", "vet", "./...")
}

func Run() error {
	mg.Deps(Build)
	args := os.Args[2:] // Get args after "mage run"
	binary := "bin/zzk"
	if runtime.GOOS == "windows" {
		binary = "bin/zzk.exe"
	}
	return sh.Run(binary, args...)
}

func VSCode() error {
	fmt.Println("Generating VS Code debug configuration...")

	if err := os.MkdirAll(".vscode", 0755); err != nil {
		return fmt.Errorf("failed to create .vscode directory: %w", err)
	}

	launchConfig := map[string]any{
		"version": "0.2.0",
		"configurations": []map[string]any{
			{
				"name":    "Debug zzk",
				"type":    "go",
				"request": "launch",
				"mode":    "debug",
				"program": "${workspaceFolder}",
				"args":    []string{"--tmp", "yt", "aud", "https://www.youtube.com/watch?v=9bZkp7q19f0"},
			},
		},
	}

	launchPath := ".vscode/launch.json"
	jsonData, err := json.MarshalIndent(launchConfig, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal launch config: %w", err)
	}

	if err := os.WriteFile(launchPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write launch.json: %w", err)
	}

	fmt.Printf("✓ Generated %s\n", launchPath)
	fmt.Println("\nTo debug your app:")
	fmt.Println("  1. Open VS Code")
	fmt.Println("  2. Set breakpoints by clicking left of line numbers")
	fmt.Println("  3. Press F5 (or go to Run and Debug)")
	fmt.Println("  4. Edit the 'args' in launch.json to pass arguments")
	fmt.Println("     Example: \"args\": [\"yt\", \"aud\", \"https://...\"]")
	fmt.Println()
	fmt.Println("That's it! VS Code will build, run, and stop at your breakpoints.")

	return nil
}
