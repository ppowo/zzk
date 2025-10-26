package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

var macosVolCmd = &cobra.Command{
	Use:   "vol [volume]",
	Short: "Set system volume to default or specified level",
	Long: `Set system volume to default (17) or specified level (0-100).

This command only works on macOS and uses AppleScript to set the volume.

Examples:
  zzk macos vol       # Set volume to default (17)
  zzk macos vol 50    # Set volume to 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "darwin" {
			return fmt.Errorf("this command only works on macOS")
		}

		volume := 17 // Default volume
		isDefault := true

		if len(args) > 0 {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("volume must be a number")
			}
			if v < 0 || v > 100 {
				return fmt.Errorf("volume must be between 0 and 100")
			}
			volume = v
			isDefault = false
		}

		return setVolume(volume, isDefault)
	},
}

func init() {
	macosCmd.AddCommand(macosVolCmd)
}

func setVolume(volume int, isDefault bool) error {
	script := fmt.Sprintf("set volume output volume %d", volume)
	cmd := exec.Command("osascript", "-e", script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error setting volume: %w", err)
	}

	if isDefault {
		fmt.Printf("Volume set to %d (default)\n", volume)
	} else {
		fmt.Printf("Volume set to %d\n", volume)
	}

	return nil
}
