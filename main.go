package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"tray-power-app/assets"
	"tray-power-app/ppd"

	"fyne.io/systray"
)

var dbusClient *ppd.Client

func main() {
	//connect to dbus
	var err error
	dbusClient, err = ppd.Connect()
	if err != nil {
		log.Fatal(err)
	}

	onExit := func() {
		if dbusClient != nil {
			_ = dbusClient.Close()
		}
		fmt.Println("Exit application")
		os.Exit(0)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	time.Sleep(100 * time.Millisecond) // Ensure tray initialization

	systray.SetTooltip("Power Profiles Daemon mini Tray applet")

	active, err := dbusClient.ActiveProfile()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("ActiveProfile: %s\n\n", active)

	profiles, err := dbusClient.Profiles()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available profiles:")
	for _, m := range profiles {
		fmt.Printf("  - %v (driver=%v)\n", m["Profile"], m["Driver"])
	}

	// Create menu items dynamically from available profiles
	profileItems := make([]*systray.MenuItem, 0, len(profiles))

	for _, m := range profiles {
		profileName := m["Profile"].Value().(string)
		profileItem := systray.AddMenuItemCheckbox(profileName, fmt.Sprintf("Set %s power profile", profileName), active == profileName)
		profileIcon := getIconForProfile(profileName)
		if active == profileName {
			systray.SetIcon(profileIcon)
			fmt.Printf("SetSystray icon: %s\n", profileName)
		}

		profileItem.SetIcon(profileIcon)
		profileItems = append(profileItems, profileItem)
		go func(profileName string, menuItem *systray.MenuItem) {
			for range menuItem.ClickedCh {
				if dbusClient == nil {
					log.Printf("DBus client not ready")
					continue
				}
				if err := dbusClient.SetActiveProfile(profileName); err != nil {
					log.Printf("Failed to set active profile to %s: %v", profileName, err)
				} else {
					for _, otherMenuItem := range profileItems {
						otherMenuItem.Uncheck()
					}
					menuItem.Check()
					systray.SetIcon(getIconForProfile(profileName))
				}
			}
		}(profileName, profileItem)
	}

	// Add separator
	systray.AddSeparator()

	// Create the quit menu item first (this is important)
	quitItem := systray.AddMenuItem("Quit", "Exit application")

	// Handle quit menu item
	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()
}

func getIconForProfile(profileName string) []byte {
	if profileIcon, ok := assets.Images[profileName]; ok {
		return profileIcon
	} else {
		//default icon
		return assets.Images["balanced"]
	}
}
