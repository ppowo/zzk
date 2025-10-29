//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// SetVolume sets the system volume to the specified level (0-100)
// Returns the previous volume level
func SetVolume(volume int) (int, error) {
	var previousVolume int

	// Get current volume before setting new one
	current, err := GetVolume()
	if err == nil {
		previousVolume = current
	} else {
		previousVolume = -1 // Unknown
	}

	switch runtime.GOOS {
	case "darwin":
		return previousVolume, setVolumeMacOS(volume)
	case "linux":
		return previousVolume, setVolumeLinux(volume)
	default:
		return previousVolume, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetVolume returns the current system volume (0-100) or -1 if unknown
func GetVolume() (int, error) {
	switch runtime.GOOS {
	case "darwin":
		return getVolumeMacOS()
	case "linux":
		return getVolumeLinux()
	default:
		return -1, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func setVolumeMacOS(volume int) error {
	script := fmt.Sprintf("set volume output volume %d", volume)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func getVolumeMacOS() (int, error) {
	script := "get output volume of (get volume settings)"
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	volume, err := strconv.Atoi(string(output[:len(output)-1])) // Trim newline
	if err != nil {
		return 0, err
	}

	return volume, nil
}

func setVolumeLinux(volume int) error {
	// Detect available audio system
	audioSystem, err := detectLinuxAudioSystem()
	if err != nil {
		return fmt.Errorf("no supported audio system found: %w", err)
	}

	var cmd *exec.Cmd
	switch audioSystem {
	case "pulseaudio":
		cmd = exec.Command("pactl", "set-sink-volume", "0", fmt.Sprintf("%d%%", volume))
	case "alsa":
		cmd = exec.Command("amixer", "set", "Master", fmt.Sprintf("%d%%", volume))
	default:
		return fmt.Errorf("unknown audio system: %s", audioSystem)
	}

	return cmd.Run()
}

func getVolumeLinux() (int, error) {
	// Detect available audio system
	audioSystem, err := detectLinuxAudioSystem()
	if err != nil {
		return 0, fmt.Errorf("no supported audio system found: %w", err)
	}

	var cmd *exec.Cmd
	var output []byte
	var parseErr error

	switch audioSystem {
	case "pulseaudio":
		cmd = exec.Command("pactl", "get-sink-volume", "0")
		output, parseErr = cmd.Output()
		if parseErr != nil {
			return 0, parseErr
		}
		// Parse output like "Volume: front-left: 50% ..."
		return parseLinuxPulseVolume(string(output))
	case "alsa":
		cmd = exec.Command("amixer", "get", "Master")
		output, parseErr = cmd.Output()
		if parseErr != nil {
			return 0, parseErr
		}
		// Parse output like "Simple mixer control 'Master',0"
		return parseLinuxAlsaVolume(string(output))
	default:
		return 0, fmt.Errorf("unknown audio system: %s", audioSystem)
	}
}

func detectLinuxAudioSystem() (string, error) {
	// Check for PulseAudio
	if isCommandAvailable("pactl") {
		// Verify PulseAudio is actually running
		cmd := exec.Command("pactl", "info")
		if err := cmd.Run(); err == nil {
			return "pulseaudio", nil
		}
	}

	// Check for ALSA
	if isCommandAvailable("amixer") {
		return "alsa", nil
	}

	return "", fmt.Errorf("no supported audio control command found (pactl, amixer)")
}

func parseLinuxPulseVolume(output string) (int, error) {
	// Parse format: "Volume: front-left: 50% ..."
	// Simple implementation: look for the first percentage
	for i := 0; i < len(output)-1; i++ {
		if output[i] == '%' {
			// Go back to find the number
			start := i
			for start > 0 && (output[start-1] >= '0' && output[start-1] <= '9') {
				start--
			}
			volStr := output[start:i]
			vol, err := strconv.Atoi(volStr)
			if err == nil {
				return vol, nil
			}
		}
	}
	return -1, fmt.Errorf("failed to parse pulseaudio volume")
}

func parseLinuxAlsaVolume(output string) (int, error) {
	// Parse format looking for percentage values
	// Look for patterns like " [XX%]"
	for i := 0; i < len(output)-3; i++ {
		if output[i] == ' ' && output[i+1] == '[' && output[i+3] == '%' {
			if output[i+2] >= '0' && output[i+2] <= '9' {
				vol := int(output[i+2] - '0')
				// Check for two-digit percentage
				if i+4 < len(output) && output[i+4] >= '0' && output[i+4] <= '9' {
					vol = vol*10 + int(output[i+4]-'0')
				}
				return vol, nil
			}
		}
	}
	return -1, fmt.Errorf("failed to parse alsa volume")
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}