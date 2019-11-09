// mute executes programs suppressing std streams if required
package main

import (
	"fmt"
	"github.com/farzadghanei/mute"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "version %v. Usage: %v COMMAND", mute.Version, os.Args[0])
		os.Exit(mute.EXEC_ERR)
	}
	// use config file if readable, otherwise use a default conf
	// to mute zero exit codes
	confPath := "/etc/mute.toml"
	envConf, envSet := os.LookupEnv("MUTE_CONFIG")
	if envSet {
		confPath = envConf
	}
	conf, err := mute.ReadConfFile(confPath)
	if err != nil {
		conf = *mute.DefaultConf()
	}
	exitCode := mute.Exec(os.Args[1], os.Args[2:], &conf, os.Stdout)
	os.Exit(exitCode)
}
