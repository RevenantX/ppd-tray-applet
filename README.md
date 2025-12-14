# PPD Tray Applet 
![Balanced Profile](assets/balanced.png)

A simple tray applet for [power-profiles-daemon](https://github.com/power-profiles-daemon/power-profiles-daemon) to manage power profiles from your system tray.

## Features

- View and switch between available power profiles (balanced, performance, powersave)
- System tray icon that updates based on the active profile

## Installation

### From Source

1. Clone this repository:
   ```bash
   git clone https://github.com/RevenantX/ppd-tray-applet.git
   cd ppd-tray-applet
   ```

2. Build the application:
   ```bash
   go build -o ppd-tray-applet
   ```

3. Run the applet:
   ```bash
   ./ppd-tray-applet
   ```

### From AUR (Arch Linux)

For Arch Linux users, you can install from the AUR:

```bash
yay -S ppd-tray-applet-git
```

or

```bash
paru -S ppd-tray-applet-git
```

## Usage

1. Launch the application (it will appear in your system tray)
2. Click on the tray icon to open the menu
3. Select a power profile from the list:
   - The currently active profile will be checked
   - Icons represent each profile type
4. To quit, select "Quit" from the menu

## Dependencies

- Go 1.21.0 or higher
- power-profiles-daemon (must be running)
- System tray support (works with most desktop environments)

## Building

To build from source:

```bash
go mod tidy
go build -o ppd-tray-applet -ldflags="-w -s"
```

## Configuration

The applet automatically detects available power profiles from power-profiles-daemon and creates menu items for each one. No manual configuration is required.

## License

MIT License - see [LICENSE](LICENSE) file for details.
