//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PowerShell-based Windows volume control
// More reliable than complex COM interface programming
func setVolumeWindows(volume int) error {
	// Use PowerShell to set volume via audio API
	// PowerShell is available on all modern Windows systems
	script := fmt.Sprintf(`
		Add-Type -TypeDefinition @"
		using System;
		using System.Runtime.InteropServices;
		public class Audio {
			[DllImport("winmm.dll")]
			public static extern int waveOutSetVolume(IntPtr hwo, uint dwVolume);

			public static void SetVolume(int volume) {
				// Convert 0-100 to 0-65535 range (both left and right channels)
				uint vol = (uint)(volume * 655.35);
				waveOutSetVolume(IntPtr.Zero, (vol << 16) | vol);
			}
		}
		"@
		[Audio]::SetVolume(%d)
	`, volume)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set volume: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func getVolumeWindows() (int, error) {
	// Use PowerShell to get current volume
	script := `
		Add-Type -TypeDefinition @"
		using System;
		using System.Runtime.InteropServices;
		public class Audio {
			[DllImport("winmm.dll")]
			public static extern int waveOutGetVolume(IntPtr hwo, out uint dwVolume);

			public static int GetVolume() {
				uint volume;
				waveOutGetVolume(IntPtr.Zero, out volume);
				// Extract volume from left channel (lower 16 bits)
				return (int)((volume & 0xFFFF) / 655.35);
			}
		}
		"@
		[Audio]::GetVolume()
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get volume: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to get the volume value
	volumeStr := strings.TrimSpace(string(output))
	volume, err := strconv.Atoi(volumeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse volume output: %s", volumeStr)
	}

	// Clamp to valid range
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}

	return volume, nil
}