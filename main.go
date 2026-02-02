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
var stopProfileWatch func()

type MenuItemInfo struct {
	menuItem *systray.MenuItem
	name     string
}

func main() {
	//connect to dbus
	var err error
	dbusClient, err = ppd.Connect()
	if err != nil {
		log.Fatal(err)
	}

	onExit := func() {
		if stopProfileWatch != nil {
			stopProfileWatch()
		}
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

	profiles, err := dbusClient.Profiles()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available profiles:")
	for _, m := range profiles {
		fmt.Printf("  - %v (driver=%v)\n", m["Profile"], m["Driver"])
	}

	profileItems := make([]MenuItemInfo, 0, len(profiles))

	updateStatus := func(activeProfile string) {
		systray.SetIcon(getIconForProfile(activeProfile))
		systray.SetTooltip(fmt.Sprintf("Power profile: %s", activeProfile))
		for _, item := range profileItems {
			if item.name == activeProfile {
				item.menuItem.Check()
			} else {
				item.menuItem.Uncheck()
			}
		}
	}

	// Create menu items dynamically from available profiles
	for _, m := range profiles {
		profileName := m["Profile"].Value().(string)
		profileItem := systray.AddMenuItemCheckbox(profileName, fmt.Sprintf("Set %s power profile", profileName), active == profileName)
		profileItem.SetIcon(getIconForProfile(profileName))
		profileItems = append(profileItems, MenuItemInfo{menuItem: profileItem, name: profileName})

		go func(profileName string, menuItem *systray.MenuItem) {
			for range menuItem.ClickedCh {
				if dbusClient == nil {
					log.Printf("DBus client not ready")
					continue
				}
				if err := dbusClient.SetActiveProfile(profileName); err != nil {
					log.Printf("Failed to set active profile to %s: %v", profileName, err)
				}
			}
		}(profileName, profileItem)
	}

	updateStatus(active)

	updates, stopWatch, err := dbusClient.SubscribeActiveProfileChanges()
	if err != nil {
		log.Printf("Failed to subscribe to ActiveProfile changes: %v", err)
	} else {
		go func() {
			for activeProfile := range updates {
				updateStatus(activeProfile)
			}
		}()
	}

	// Create the quit menu item first (this is important)
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Exit application")

	// Handle quit menu item
	go func() {
		<-quitItem.ClickedCh
		stopWatch()
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
