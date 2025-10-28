# zzk

A command-line swiss army knife with diverse functionality.

## Features

- YouTube Downloads - Download audio, albums/playlists, and videos with aria2c acceleration
- Smart Video Quality - Automatically detects your screen resolution and downloads appropriate quality
- Git Identity Management - Manage multiple git identities for different domains and folders
- Claude API Providers - Switch between different Claude API providers for Claude Code
- Backup/Restore - Backup and restore directories with automatic verification (macOS/Linux)
- Font Installation - Install custom fonts with a single command
- Volume Control - Cross-platform system volume control (macOS, Windows, Linux)
- macOS Utilities - Other macOS-specific tools
- Self-Managing - Automatically downloads and manages its own yt-dlp binary

## Installation

### Prerequisites

- **Go 1.25+** (for building from source)
- **aria2c** - Required for accelerated downloads

Install aria2c:
```bash
# macOS
brew install aria2

# Linux (Debian/Ubuntu)
sudo apt install aria2

# Linux (Fedora)
sudo dnf install aria2

# Windows
scoop install aria2
# or
choco install aria2
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/ppowo/zzk.git
cd zzk

# Install build tools
go generate

# Build using Mage
mage build

# Install to ~/.bio/bin
mage install
```

## Usage

### YouTube Downloads

```bash
# Download audio
zzk yt aud https://youtube.com/watch?v=dQw4w9WgXcQ

# Download album/playlist
zzk yt alb https://youtube.com/playlist?list=...

# Download video (smart quality based on screen resolution)
zzk yt vid https://youtube.com/watch?v=dQw4w9WgXcQ
```

Output: Audio to `~/Music/`, Videos to `~/Movies/`

### Backup and Restore

```bash
# Backup .bio directory
zzk backup bio

# Restore .bio from a code
zzk backup bio abc123
```

Available targets: `bio`, `openemu` (macOS only)

### Git Identity Management

Manage multiple git identities (user, email, SSH keys) for different domains and folders.

Create `~/.git-identities.json`:
```json
{
  "identities": [
    {
      "name": "work",
      "domain": "github.com",
      "user": "John Doe",
      "email": "john@company.com",
      "folders": ["~/work/"]
    },
    {
      "name": "personal",
      "domain": "github.com",
      "user": "johndoe",
      "email": "john@personal.com",
      "folders": ["~/personal/", "~/projects/"]
    }
  ]
}
```

Commands:
```bash
zzk git sync    # Generate SSH keys, update git config, and configure SSH
zzk git ls      # List all identities
zzk git where   # Show which identity applies to current directory
zzk git info <identity-name>  # Show detailed information about an identity
```

### Claude API Provider Management

Manage multiple Claude API providers for use with Claude Code.

One-time setup - add to shell config (`~/.zshrc`, `~/.bashrc`):
```bash
[ -f ~/.config/zzk/claude-env.sh ] && source ~/.config/zzk/claude-env.sh
```

Commands:
```bash
zzk claude add <provider-name>     # Add a new provider
zzk claude use <provider-name>     # Switch active provider
zzk claude ls                      # List all providers
zzk claude edit <provider-name>    # Edit a provider
zzk claude rm <provider-name>      # Remove a provider
zzk claude reset                   # Reset to official Anthropic API
```

### Font Installation

```bash
# Install DMCA Sans Serif font
zzk font-install dmca
```

### System Volume Control

Control system volume (cross-platform: macOS, Windows, Linux)

```bash
# Set volume to default (17)
zzk vol

# Set volume to specific level (0-100)
zzk vol 50
```

**Platform Requirements:**
- **macOS**: Uses AppleScript (built-in)
- **Windows**: Uses Windows Audio API
- **Linux**: Automatically detects PulseAudio (`pactl`) or ALSA (`amixer`)

## Development

This project uses [Mage](https://magefile.org/) for build automation.

```bash
mage build     # Build the project
mage install   # Install to ~/.bio/bin
mage test      # Run tests
mage fmt       # Format code
mage vet       # Run go vet
mage check     # Run all checks (fmt + vet + test)
mage clean     # Clean build artifacts
```

## Requirements

- **Go**: 1.25 or higher
- **OS**: macOS, Linux (X11), Windows
  - Linux requires `xrandr` for screen resolution detection
- **aria2c**: Must be installed and in PATH

## License

See [LICENSE](LICENSE) file for details.