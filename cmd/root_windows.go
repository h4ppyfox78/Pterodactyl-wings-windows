package cmd

import (
	"errors"
	"fmt"
	log2 "log"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/mitchellh/colorstring"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/system"
)

const PathValidationRegex = `(?m)^[a-zA-Z]:\\`

// Reads the configuration from the disk and then sets up the global singleton
// with all the configuration values.
func initConfig() {
	var re = regexp.MustCompile(PathValidationRegex)

	if !re.MatchString(configPath) {
		d, err := os.Getwd()

		if err != nil {
			log2.Fatalf("cmd/root: error getting current working directory: %s", err)
		}

		configPath = path.Clean(path.Join(d, configPath))
	}

	err := config.FromFile(configPath)
	if err != nil {
		d, err := os.Getwd()

		if err != nil {
			log2.Fatalf("cmd/root: error getting current working directory: %s", err)
		}

		err = config.FromFile(path.Clean(path.Join(d, "config.yml")))

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				exitWithConfigurationNotice()
			}
			log2.Fatalf("cmd/root: error while reading configuration file: %s", err)
		}
	}
	if debug && !config.Get().Debug {
		config.SetDebugViaFlag(debug)
	}
}

func exitWithConfigurationNotice() {
	fmt.Print(colorstring.Color(`
[_red_][white][bold]Error: Configuration File Not Found[reset]

Wings was not able to locate your configuration file, and therefore is not
able to complete its boot process. Please ensure you have copied your instance
configuration file into the default location below.

Default Location: C:\ProgramData\Pterodactyl\config.yml

[yellow]This is not a bug with this software. Please do not make a bug report
for this issue, it will be closed.[reset]

`))
	os.Exit(1)
}

// Prints the wings logo, nothing special here!
func printLogo() {
	fmt.Printf(colorstring.Color(`
                     
__ [blue][bold]Pterodactyl[reset] _____ _________ _______ _______ ______
\_____\    \/\/    /   __    /       /  __   /   ___/
   \___\          /   /_/   /   /   /  /_/  /___   /
        \___/\___/_________/___/___/___    /______/
                            	  /_______/ [bold]%s[reset]

Copyright © 2018 - %d Dane Everitt & Contributors

Website:  https://pterodactyl.io
 Source:  https://github.com/pterodactyl/wings
License:  https://github.com/pterodactyl/wings/blob/develop/LICENSE

This software is made available under the terms of the MIT license.
The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.%s`), system.Version, time.Now().Year(), "\n\n")
}
