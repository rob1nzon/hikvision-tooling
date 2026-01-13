package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/cameronnewman/hikvision-tooling/internal/config"
	"github.com/cameronnewman/hikvision-tooling/internal/crypto"
	"github.com/cameronnewman/hikvision-tooling/internal/logger"
	"github.com/cameronnewman/hikvision-tooling/internal/network"
	"github.com/cameronnewman/hikvision-tooling/internal/sadp"
)

// App represents the GUI application
type App struct {
	app      fyne.App
	window   fyne.Window
	config   *config.Config
	logger   *logger.Logger
	devices  []*sadp.Device
	scanner  *sadp.Scanner
	deviceList *widget.List
	statusLabel *widget.Label
	detailsText *widget.Label
	selectedDevice *sadp.Device
}

// NewApp creates a new GUI application
func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log := logger.New(cfg.Debug)

	a := &App{
		app:     app.NewWithID("com.hikvision.tooling"),
		config:  cfg,
		logger:  log,
		devices: make([]*sadp.Device, 0),
		scanner: sadp.NewScanner(cfg.SADPTimeout, log),
	}

	a.app.SetIcon(theme.ComputerIcon())
	a.window = a.app.NewWindow("Hikvision Camera Manager")
	a.window.Resize(fyne.NewSize(1000, 600))
	a.window.CenterOnScreen()

	return a
}

// Run starts the GUI application
func (a *App) Run() {
	a.buildUI()
	a.window.ShowAndRun()
}

func (a *App) buildUI() {
	// Create toolbar
	toolbar := a.createToolbar()

	// Create device list
	deviceListContainer := a.createDeviceList()

	// Create details panel
	detailsPanel := a.createDetailsPanel()

	// Create status bar
	a.statusLabel = widget.NewLabel("Ready")

	// Split layout: device list on left, details on right
	split := container.NewHSplit(
		deviceListContainer,
		detailsPanel,
	)
	split.Offset = 0.4

	// Main layout
	content := container.NewBorder(
		toolbar,
		a.statusLabel,
		nil,
		nil,
		split,
	)

	a.window.SetContent(content)
}

func (a *App) createToolbar() *fyne.Container {
	// Scan button with dropdown
	scanBtn := widget.NewButton("Scan SADP", func() {
		a.scanSADP()
	})
	scanBtn.Icon = theme.ViewRefreshIcon()

	// Scan ARP button
	scanARPBtn := widget.NewButton("Scan Network", func() {
		a.showARPScanDialog()
	})
	scanARPBtn.Icon = theme.SearchIcon()

	// Reset password button
	resetBtn := widget.NewButton("Reset Password", func() {
		a.showResetDialog()
	})
	resetBtn.Icon = theme.ContentClearIcon()

	// About button
	aboutBtn := widget.NewButton("About", func() {
		a.showAbout()
	})
	aboutBtn.Icon = theme.InfoIcon()

	return container.NewHBox(
		scanBtn,
		scanARPBtn,
		widget.NewSeparator(),
		resetBtn,
		layout.NewSpacer(),
		aboutBtn,
	)
}

func (a *App) createDeviceList() *fyne.Container {
	title := widget.NewLabelWithStyle("Discovered Devices", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.deviceList = widget.NewList(
		func() int {
			return len(a.devices)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(a.devices) {
				dev := a.devices[id]
				box := item.(*fyne.Container)
				
				ipLabel := box.Objects[0].(*widget.Label)
				ipLabel.SetText(fmt.Sprintf("%s - %s", dev.IPv4Address, dev.DeviceType))
				ipLabel.TextStyle = fyne.TextStyle{Bold: true}
				
				infoLabel := box.Objects[1].(*widget.Label)
				status := "Inactive"
				if dev.Activated == "true" {
					status = "Active"
				}
				infoLabel.SetText(fmt.Sprintf("MAC: %s | Status: %s", dev.MAC, status))
			}
		},
	)

	a.deviceList.OnSelected = func(id widget.ListItemID) {
		if id < len(a.devices) {
			a.selectedDevice = a.devices[id]
			a.updateDetails()
		}
	}

	return container.NewBorder(
		title,
		nil,
		nil,
		nil,
		a.deviceList,
	)
}

func (a *App) createDetailsPanel() *fyne.Container {
	title := widget.NewLabelWithStyle("Device Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	a.detailsText = widget.NewLabel("Select a device to view details")
	a.detailsText.Wrapping = fyne.TextWrapWord

	// Configuration buttons
	activateBtn := widget.NewButton("Activate Device", func() {
		a.activateDevice()
	})
	activateBtn.Icon = theme.ConfirmIcon()

	updateIPBtn := widget.NewButton("Update IP", func() {
		a.showUpdateIPDialog()
	})
	updateIPBtn.Icon = theme.DocumentCreateIcon()

	rebootBtn := widget.NewButton("Reboot", func() {
		a.rebootDevice()
	})
	rebootBtn.Icon = theme.MediaReplayIcon()

	btnBox := container.NewHBox(
		activateBtn,
		updateIPBtn,
		rebootBtn,
	)

	scrollDetails := container.NewScroll(a.detailsText)
	scrollDetails.SetMinSize(fyne.NewSize(400, 400))

	return container.NewBorder(
		title,
		btnBox,
		nil,
		nil,
		scrollDetails,
	)
}

func (a *App) scanSADP() {
	a.statusLabel.SetText("Scanning for devices via SADP protocol...")
	a.deviceList.UnselectAll()
	a.selectedDevice = nil
	a.updateDetails()

	go func() {
		devices, err := a.scanner.Discover()
		if err != nil {
			a.statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			dialog.ShowError(err, a.window)
			return
		}

		a.devices = devices
		a.deviceList.Refresh()
		a.statusLabel.SetText(fmt.Sprintf("Found %d device(s)", len(devices)))
	}()
}

func (a *App) showARPScanDialog() {
	cidrEntry := widget.NewEntry()
	cidrEntry.SetPlaceHolder("192.168.1.0/24")
	cidrEntry.SetText("192.168.1.0/24")

	workersEntry := widget.NewEntry()
	workersEntry.SetPlaceHolder("100")
	workersEntry.SetText("100")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Network CIDR:", Widget: cidrEntry},
			{Text: "Workers:", Widget: workersEntry},
		},
		OnSubmit: func() {
			cidr := cidrEntry.Text
			a.scanARP(cidr)
		},
	}

	dialog.ShowForm("Scan Network (ARP)", "Scan", "Cancel", form.Items, form.OnSubmit, a.window)
}

func (a *App) scanARP(cidr string) {
	a.statusLabel.SetText(fmt.Sprintf("Scanning network %s...", cidr))

	go func() {
		ips, err := network.ExpandCIDR(cidr)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid CIDR: %w", err), a.window)
			a.statusLabel.SetText("Ready")
			return
		}

		// Get ARP table
		arpTable, err := network.GetARPTable()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read ARP table: %w", err), a.window)
			a.statusLabel.SetText("Ready")
			return
		}

		// Filter Hikvision devices
		var foundDevices []string
		for _, ip := range ips {
			if mac, ok := arpTable[ip]; ok {
				if network.IsHikvisionMAC(mac) {
					foundDevices = append(foundDevices, fmt.Sprintf("%s - %s", ip, mac))
				}
			}
		}

		a.statusLabel.SetText(fmt.Sprintf("Found %d Hikvision device(s) via ARP", len(foundDevices)))
		
		if len(foundDevices) > 0 {
			msg := "Found devices:\n\n" + strings.Join(foundDevices, "\n")
			dialog.ShowInformation("ARP Scan Results", msg, a.window)
		} else {
			dialog.ShowInformation("ARP Scan Results", "No Hikvision devices found", a.window)
		}
	}()
}

func (a *App) updateDetails() {
	if a.selectedDevice == nil {
		a.detailsText.SetText("Select a device to view details")
		return
	}

	dev := a.selectedDevice
	status := "Inactive"
	if dev.Activated == "true" {
		status = "Active"
	}

	details := fmt.Sprintf(`Device Information
==================

IP Address:        %s
MAC Address:       %s
Device Type:       %s
Serial Number:     %s
Firmware Version:  %s
Command Port:      %d
HTTP Port:         %d
Status:            %s

Network Configuration
=====================
IPv4 Address:      %s
IPv4 Subnet Mask:  %s
IPv4 Gateway:      %s
DHCP:              %s

Additional Info
===============
UUID:              %s
Boot Time:         %s
`,
		dev.IPv4Address,
		dev.MAC,
		dev.DeviceType,
		dev.DeviceSN,
		dev.SoftwareVersion,
		dev.CommandPort,
		dev.HttpPort,
		status,
		dev.IPv4Address,
		dev.IPv4SubnetMask,
		dev.IPv4Gateway,
		dev.DHCP,
		dev.Uuid,
		dev.BootTime,
	)

	a.detailsText.SetText(details)
}

func (a *App) activateDevice() {
	if a.selectedDevice == nil {
		dialog.ShowInformation("No Device", "Please select a device first", a.window)
		return
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("New password")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password:", Widget: passwordEntry},
		},
		OnSubmit: func() {
			a.sendActivateCommand(passwordEntry.Text)
		},
	}

	dialog.ShowForm("Activate Device", "Activate", "Cancel", form.Items, form.OnSubmit, a.window)
}

func (a *App) sendActivateCommand(password string) {
	if a.selectedDevice == nil {
		return
	}

	dev := a.selectedDevice
	opts := sadp.SendOptions{
		TargetIP:  dev.IPv4Address,
		TargetMAC: dev.MAC,
		Password:  password,
		Timeout:   a.config.SADPTimeout,
	}

	go func() {
		response, err := a.scanner.SendCommand("activate", opts)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		dialog.ShowInformation("Response", response, a.window)
	}()
}

func (a *App) showUpdateIPDialog() {
	if a.selectedDevice == nil {
		dialog.ShowInformation("No Device", "Please select a device first", a.window)
		return
	}

	dev := a.selectedDevice

	ipEntry := widget.NewEntry()
	ipEntry.SetText(dev.IPv4Address)

	maskEntry := widget.NewEntry()
	maskEntry.SetText(dev.IPv4SubnetMask)

	gatewayEntry := widget.NewEntry()
	gatewayEntry.SetText(dev.IPv4Gateway)

	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", dev.CommandPort))

	passwordEntry := widget.NewPasswordEntry()

	dhcpCheck := widget.NewCheck("Enable DHCP", nil)
	dhcpCheck.Checked = dev.DHCP == "true"

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "New IP:", Widget: ipEntry},
			{Text: "Subnet Mask:", Widget: maskEntry},
			{Text: "Gateway:", Widget: gatewayEntry},
			{Text: "Port:", Widget: portEntry},
			{Text: "DHCP:", Widget: dhcpCheck},
			{Text: "Password:", Widget: passwordEntry},
		},
		OnSubmit: func() {
			a.updateDeviceIP(ipEntry.Text, maskEntry.Text, gatewayEntry.Text, portEntry.Text, dhcpCheck.Checked, passwordEntry.Text)
		},
	}

	dialog.ShowForm("Update Device IP", "Update", "Cancel", form.Items, form.OnSubmit, a.window)
}

func (a *App) updateDeviceIP(newIP, mask, gateway, port string, dhcp bool, password string) {
	if a.selectedDevice == nil {
		return
	}

	dev := a.selectedDevice
	opts := sadp.SendOptions{
		TargetIP:   dev.IPv4Address,
		TargetMAC:  dev.MAC,
		Password:   password,
		NewIP:      newIP,
		NewMask:    mask,
		NewGateway: gateway,
		DHCP:       dhcp,
		Timeout:    a.config.SADPTimeout,
	}

	go func() {
		response, err := a.scanner.SendCommand("update", opts)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		dialog.ShowInformation("Response", response, a.window)
	}()
}

func (a *App) rebootDevice() {
	if a.selectedDevice == nil {
		dialog.ShowInformation("No Device", "Please select a device first", a.window)
		return
	}

	passwordEntry := widget.NewPasswordEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password:", Widget: passwordEntry},
		},
		OnSubmit: func() {
			a.sendRebootCommand(passwordEntry.Text)
		},
	}

	dialog.ShowForm("Reboot Device", "Reboot", "Cancel", form.Items, form.OnSubmit, a.window)
}

func (a *App) sendRebootCommand(password string) {
	if a.selectedDevice == nil {
		return
	}

	dev := a.selectedDevice
	opts := sadp.SendOptions{
		TargetIP:  dev.IPv4Address,
		TargetMAC: dev.MAC,
		Password:  password,
		Timeout:   a.config.SADPTimeout,
	}

	go func() {
		response, err := a.scanner.SendCommand("reboot", opts)
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		dialog.ShowInformation("Response", response, a.window)
	}()
}

func (a *App) showResetDialog() {
	serialEntry := widget.NewEntry()
	serialEntry.SetPlaceHolder("Device serial (without model prefix)")

	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("YYYYMMDD")
	dateEntry.SetText(time.Now().Format("20060102"))

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	calculateBtn := widget.NewButton("Calculate", func() {
		serial := serialEntry.Text
		date := dateEntry.Text

		if serial == "" || date == "" {
			dialog.ShowError(fmt.Errorf("serial and date are required"), a.window)
			return
		}

		// Import crypto package and generate reset code
		resetCode := a.generateResetCode(serial, date)
		resultLabel.SetText(fmt.Sprintf("Reset Code: %s\n\nNote: This only works on firmware < 5.3.0", resetCode))
	})

	content := container.NewVBox(
		widget.NewLabel("Password Reset Code Generator"),
		widget.NewSeparator(),
		widget.NewForm(
			&widget.FormItem{Text: "Serial Number:", Widget: serialEntry},
			&widget.FormItem{Text: "Date (YYYYMMDD):", Widget: dateEntry},
		),
		calculateBtn,
		widget.NewSeparator(),
		resultLabel,
	)

	dialog.ShowCustom("Reset Password", "Close", content, a.window)
}

func (a *App) generateResetCode(serial, date string) string {
	return crypto.GenerateResetCode(serial, date)
}

func (a *App) showAbout() {
	content := widget.NewLabel(`Hikvision Camera Manager
Version 1.0.0

A cross-platform GUI tool for discovering and
managing Hikvision devices on your network.

Features:
• SADP protocol device discovery
• ARP-based network scanning
• Device activation and configuration
• Password reset code generation
• Device reboot functionality

Built with Fyne UI framework
`)
	content.Wrapping = fyne.TextWrapWord

	dialog.ShowCustom("About", "Close", content, a.window)
}
