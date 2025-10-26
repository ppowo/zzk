package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ppowo/zzk/internal/fileutil"
	"github.com/ppowo/zzk/internal/font"
	"github.com/spf13/cobra"
)

var fontInstallDmcaCmd = &cobra.Command{
	Use:   "dmca",
	Short: "Install DMCA Sans Serif font",
	Long: `Install DMCA Sans Serif font to user font directory (no admin/sudo required).

The font will be downloaded from the official source and installed to:
  - macOS: ~/Library/Fonts
  - Linux: ~/.local/share/fonts
  - Windows: %LOCALAPPDATA%\Microsoft\Windows\Fonts

After installation, you may need to restart applications to use the new font.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installDMCAFont()
	},
}

func init() {
	fontInstallCmd.AddCommand(fontInstallDmcaCmd)
}

func installDMCAFont() error {
	// Get user font directory based on OS
	fontDir, err := font.GetUserFontDir()
	if err != nil {
		return fmt.Errorf("failed to get font directory: %w", err)
	}

	// Create font directory if it doesn't exist
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %w", err)
	}

	// Create temporary directory for download and extraction
	tempDir, err := os.MkdirTemp("", "zzk-dmca-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download font
	fmt.Println("Downloading DMCA Sans Serif font...")
	zipPath := filepath.Join(tempDir, "DMCAsansserif9.0-20252.zip")
	if err := downloadFile(zipPath, "https://typedesign.replit.app/DMCAsansserif9.0-20252.zip"); err != nil {
		return fmt.Errorf("failed to download font: %w", err)
	}

	// Extract zip file
	fmt.Println("Extracting font files...")
	if err := unzip(zipPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}

	// Install fonts (copy TTF files to font directory)
	fmt.Printf("Installing DMCA Sans Serif to %s...\n", fontDir)
	installedCount := 0
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".ttf" {
			src := filepath.Join(tempDir, entry.Name())
			dst := filepath.Join(fontDir, entry.Name())
			if err := fileutil.CopyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy font file %s: %w", entry.Name(), err)
			}
			installedCount++
		}
	}

	if installedCount == 0 {
		return fmt.Errorf("no TTF font files found in archive")
	}

	// Refresh font cache (best effort, platform-specific)
	font.RefreshFontCache()

	fmt.Printf("✓ Successfully installed %d font file(s)!\n", installedCount)
	fmt.Println("You may need to restart applications to use the new font.")

	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Protect against zip slip
		fpath := filepath.Join(dest, f.Name)
		cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
		cleanPath := filepath.Clean(fpath)
		if !strings.HasPrefix(cleanPath, cleanDest) && cleanPath != filepath.Clean(dest) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

