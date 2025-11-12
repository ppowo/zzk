package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ytAudCmd = &cobra.Command{
	Use:   "aud [URL...]",
	Short: "Download audio from YouTube URL(s)",
	Long:  `Downloads audio from the provided URL(s) to ~/Music using yt-dlp with aria2c.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := CheckAria2c(); err != nil {
			return err
		}
		if err := CheckYtDlp(); err != nil {
			return err
		}
		var destDir string
		if UseTmpDir {
			destDir = filepath.Join(os.TempDir(), "zzk-debug")
		} else {
			destDir = filepath.Join(os.Getenv("HOME"), "Music")
		}
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}
		if err := os.Chdir(destDir); err != nil {
			return fmt.Errorf("failed to change to directory %s: %w", destDir, err)
		}
		fmt.Printf("Downloading audio to: %s\n", destDir)
		cmdArgs := append(GetAudioArgs(), args...)
		ytCmd := exec.Command("yt-dlp", cmdArgs...)
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
	ytCmd.AddCommand(ytAudCmd)
}
