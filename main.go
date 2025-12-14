package main

import (
	"fmt"
	"log"
	"os"
	"tray-power-app/assets"
	"tray-power-app/ppd"

	"fyne.io/systray"
)

func main() {
	onExit := func() {
		fmt.Println("Exit application")
		os.Exit(0)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTooltip("Power Profiles Daemon mini Tray applet")

	//connect to dbus
	client, err := ppd.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	active, err := client.ActiveProfile()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("ActiveProfile: %s\n\n", active)

	profiles, err := client.Profiles()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Available profiles:")
	for _, m := range profiles {
		profile := m["Profile"]
		driver := m["Driver"]
		fmt.Printf("  - %v (driver=%v)\n", profile.Value(), driver.Value())
	}

	// Create menu items dynamically from available profiles
	var profileItems []*systray.MenuItem

	// Store profile names to properly map clicks back to the right profile
	profileNames := make([]string, len(profiles))

	for i, m := range profiles {
		profileName := m["Profile"].Value().(string)
		profileNames[i] = profileName
		profileItem := systray.AddMenuItemCheckbox(profileName, fmt.Sprintf("Set %s power profile", profileName), active == profileName)
		if profileIcon, ok := assets.Images[profileName]; ok {
			profileItem.SetIcon(profileIcon)
			if active == profileName {
				systray.SetIcon(profileIcon)
			}
		}
		profileItems = append(profileItems, profileItem)
	}

	// Add separator
	systray.AddSeparator()

	// Create the quit menu item first (this is important)
	quitItem := systray.AddMenuItem("Quit", "Exit application")

	// Handle menu item clicks with individual goroutines for each item
	for i, item := range profileItems {
		go func(profileName string, menuItem *systray.MenuItem) {
			for range menuItem.ClickedCh {
				var client *ppd.Client
				var err error
				client, err = ppd.Connect()
				if err != nil {
					log.Printf("Failed to connect: %v", err)
					continue
				}

				//fmt.Printf("Setting active profile to: %s\n", profileName)

				err = client.SetActiveProfile(profileName)
				client.Close()

				if err != nil {
					log.Printf("Failed to set active profile to %s: %v", profileName, err)
				} else {
					for _, otherMenuItem := range profileItems {
						otherMenuItem.Uncheck()
					}
					menuItem.Check()

					if profileIcon, ok := assets.Images[profileName]; ok {
						systray.SetIcon(profileIcon)
					}
				}
			}
		}(profileNames[i], item)
	}

	// Handle quit menu item
	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()
}
