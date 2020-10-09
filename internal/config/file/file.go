package file

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/eiannone/keyboard"
)

func Create(contents string) {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".tfdr")
	fileName := "config.yaml"
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				log.Fatalf("Unable to create config directiory in path: %s. Error: %s", configDir, err.Error())
			}
		}
	}
	cfgFile := filepath.Join(configDir, fileName)
	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		saveConfig(cfgFile, contents)
	} else {
		// reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Config file (%s) found, Overwrite? [Y/n] ", cfgFile)
		txt, key, _ := keyboard.GetSingleKey()
		if key == keyboard.KeyEnter || txt == 'Y' || txt == 'y' {
			saveConfig(cfgFile, contents)
		}
	}
}

func saveConfig(cfgFile string, contents string) {
	file, err := os.Create(cfgFile)
	if err != nil {
		log.Fatalf("Error: failed while creating config file. Error: %s", err.Error())
	}
	defer file.Close()

	_, err = io.WriteString(file, contents)
	if err != nil {
		log.Fatal("Error: failed while attempting to write config yaml")
	}
	_ = file.Sync()

	fmt.Println("\nSuccessfully configured terraform disaster recovery script. Use `tfdr config get` to view your configuration.")
}
