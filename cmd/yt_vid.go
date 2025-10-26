package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ytVidCmd = &cobra.Command{
	Use:   "vid [URL...]",
	Short: "Download video from YouTube URL(s)",
	Long:  `Downloads video from the provided URL(s) to ~/Movies using yt-dlp with aria2c.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckAria2c(); err != nil {
			return err
		}
		if err := EnsureYtDlp(); err != nil {
			return fmt.Errorf("failed to ensure yt-dlp: %w", err)
		}
		var destDir string
		if UseTmpDir {
			destDir = filepath.Join(os.TempDir(), "zzk-debug")
		} else {
			destDir = filepath.Join(os.Getenv("HOME"), "Movies")
		}
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}
		if err := os.Chdir(destDir); err != nil {
			return fmt.Errorf("failed to change to directory %s: %w", destDir, err)
		}
		fmt.Printf("Downloading video to: %s\n", destDir)

		ytDlpPath := GetYtDlpPath()
		videoArgs, err := GetVideoArgs()
		if err != nil {
			return fmt.Errorf("failed to get video args: %w", err)
		}
		cmdArgs := append(videoArgs, args...)

		ytCmd := exec.Command(ytDlpPath, cmdArgs...)
		ytCmd.Stdout = os.Stdout
		ytCmd.Stderr = os.Stderr
		if err := ytCmd.Run(); err != nil {
			return fmt.Errorf("yt-dlp failed: %w", err)
		}
		fmt.Println("✓ Download completed successfully!")
		return nil
	},
}

func init() {
	ytCmd.AddCommand(ytVidCmd)
}
