# Hikvision Camera Manager GUI

A cross-platform graphical user interface for discovering and managing Hikvision cameras on your network.

## Features

- **SADP Protocol Discovery**: Discover devices using the official SADP multicast protocol
- **ARP Network Scanning**: Find devices by scanning IP ranges
- **Device List View**: View all discovered devices with detailed information
- **Device Configuration**: Activate devices, update IP settings, and reboot devices
- **Password Reset Tool**: Generate password reset codes for older firmware versions (< 5.3.0)
- **Real-time Status**: Monitor device status and connectivity

## Installation

### From Pre-built Binaries

Download the appropriate binary for your platform from the [Releases](https://github.com/cameronnewman/hikvision-tooling/releases) page:

- **Linux**: `sadp-gui-linux-amd64`
- **Windows**: `sadp-gui-windows-amd64.exe`

### From Source

#### Prerequisites

- Go 1.21 or later
- For Linux: libgl1-mesa-dev, xorg-dev
- For Windows cross-compilation: gcc-mingw-w64-x86-64

#### Build Commands

```bash
# Build for current platform
make build-gui

# Build for Windows and Linux
make release-gui

# Run the GUI
make run-gui
```

## Usage

### Main Window

The application features a split-pane interface:

**Left Pane - Device List**
- Shows all discovered Hikvision devices
- Displays IP address, device type, MAC address, and status
- Click on a device to view detailed information

**Right Pane - Device Details**
- Shows comprehensive device information
- Provides action buttons for device management

### Toolbar Actions

**Scan SADP**
- Discovers devices using the SADP multicast protocol
- No configuration required
- Automatically finds devices on the local network

**Scan Network**
- Scans a specific IP range using ARP
- Enter a CIDR notation (e.g., 192.168.1.0/24)
- Configure the number of concurrent workers

**Reset Password**
- Generate password reset codes
- Enter device serial number and date
- Works only on firmware versions < 5.3.0

### Device Management

**Activate Device**
- Set initial password for new/inactive devices
- Required before configuring the device

**Update IP**
- Change device IP address settings
- Configure subnet mask and gateway
- Enable/disable DHCP
- Requires admin password

**Reboot**
- Remotely reboot the device
- Requires admin password

## Screenshots

### Main Interface
The main window shows the device list on the left and device details on the right.

### SADP Discovery
Click "Scan SADP" to automatically discover all Hikvision devices on your network.

### Network Scanning
Enter a CIDR range to scan for devices using ARP.

### Device Configuration
Select a device and use the action buttons to configure it.

## System Requirements

### Linux
- Ubuntu 18.04+ / Debian 10+
- X11 or Wayland display server
- OpenGL support

### Windows
- Windows 10 or later
- OpenGL support (typically built-in)

## Network Requirements

- The application must be run on the same network as the Hikvision devices
- For SADP discovery: UDP port 37020 must be accessible
- For device management: HTTP/HTTPS ports (typically 80/443 or 8000) must be accessible

## Troubleshooting

### No Devices Found (SADP)

1. Ensure devices are powered on and connected to the network
2. Check that multicast traffic is not blocked by firewall
3. Verify you're on the same subnet as the devices

### No Devices Found (ARP)

1. Ensure the CIDR range includes your devices
2. Check ARP table access permissions (may require root/admin on some systems)
3. Try pinging devices first to populate ARP cache

### Cannot Configure Device

1. Verify the device password is correct
2. Check network connectivity to the device
3. Ensure device is not in "inactive" state (activate first)

### Build Errors

**Linux**: Install required dependencies
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
```

**Windows Cross-compilation**: Install MinGW
```bash
sudo apt-get install gcc-mingw-w64-x86-64
```

## Architecture

The GUI is built using:

- **[Fyne](https://fyne.io/)**: Cross-platform UI framework written in Go
- **Go 1.21+**: Programming language
- **Existing SADP Library**: Reuses the CLI tool's protocol implementation

The GUI shares the same core libraries as the CLI tool:
- `internal/sadp`: SADP protocol implementation
- `internal/network`: Network utilities and ARP scanning
- `internal/crypto`: Password reset code generation
- `internal/config`: Configuration management

## Development

### Project Structure

```
cmd/sadp-gui/
  └── main.go           # GUI entry point
internal/gui/
  └── app.go            # Main application logic and UI
```

### Running in Development

```bash
# Run directly (requires GUI environment)
make run-gui

# Build and test
make build-gui
./bin/sadp-gui
```

### Adding New Features

1. Add new UI components in `internal/gui/app.go`
2. Reuse existing core functionality from `internal/` packages
3. Update this README with new feature documentation

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Related Tools

- **CLI Tool**: [`sadp`](README.md) - Command-line interface for scripting and automation
- **Official SADP Tool**: Hikvision's Windows-only tool (this is a cross-platform alternative)

## Disclaimer

This tool is intended for legitimate network administration and security research purposes. Only use it on networks and devices you own or have explicit permission to test. The authors are not responsible for any misuse of this tool.
