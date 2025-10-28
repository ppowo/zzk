package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var volCmd = &cobra.Command{
	Use:   "vol [volume]",
	Short: "Set system volume to default or specified level",
	Long: `Set system volume to default (17) or specified level (0-100).

Examples:
  zzk vol       # Set volume to default (17)
  zzk vol 50    # Set volume to 50`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		previousVolume, err := SetVolume(volume)
		if err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}

		if isDefault {
			if previousVolume >= 0 {
				fmt.Printf("Volume set to %d (default, was %d)\n", volume, previousVolume)
			} else {
				fmt.Printf("Volume set to %d (default)\n", volume)
			}
		} else {
			if previousVolume >= 0 {
				fmt.Printf("Volume set to %d (was %d)\n", volume, previousVolume)
			} else {
				fmt.Printf("Volume set to %d\n", volume)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(volCmd)
}