// scw interract with Scaleway from the command line
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

// Command is a Scaleway command
type Command struct {
	// Exec executes the command
	Exec func(cmd *Command, args []string)

	// Usage is the one-line usage message.
	UsageLine string

	// Description is the description of the command
	Description string

	// Help is the full description of the command
	Help string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Name returns the command's name
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Options returns a string describing options of the command
func (c *Command) Options() string {
	var options string
	visitor := func(flag *flag.Flag) {
		options += fmt.Sprintf("  -%-12s %s (%s)\n", flag.Name, flag.Usage, flag.DefValue)
	}
	c.Flag.VisitAll(visitor)
	if len(options) == 0 {
		options = ""
	}
	return options
}

var commands = []*Command{
	cmdHelp,
	cmdLogin,
	cmdPs,
	cmdStart,
	cmdVersion,
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	name := args[0]
	args = args[1:]

	for _, cmd := range commands {
		if cmd.Name() == name {
			cmd.Flag.SetOutput(ioutil.Discard)
			err := cmd.Flag.Parse(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "usage: scw %s\n", cmd.UsageLine)
				os.Exit(1)
			}
			cmd.Exec(cmd, cmd.Flag.Args())
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "scw: unknown subcommand %s\nRun 'scw help for usage.\n", name)
	os.Exit(1)
}

func usage() {
	cmdHelp.Exec(cmdHelp, []string{})
	os.Exit(1)
}

// Config is a Scaleway CLI configuration file
type Config struct {
	// APIEndpoint is the endpoint to the Scaleway API
	APIEndPoint string `json:"api_endpoint"`

	// Organization is the identifier of the Scaleway orgnization
	Organization string `json:"organization"`

	// Token is the authentication token for the Scaleway organization
	Token string `json:"token"`
}

// GetConfigFilePath returns the path to the Scaleway CLI config file
func GetConfigFilePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.scwrc", u.HomeDir), nil
}

// GetConfig returns the Scaleway CLI config file for the current user
func GetConfig() (*Config, error) {
	scwrc_path, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadFile(scwrc_path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetScalewayAPI returns a ScalewayAPI using the user config file
func GetScalewayAPI() (*ScalewayAPI, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}
	api := NewScalewayAPI(config.APIEndPoint, config.Organization, config.Token)
	return api, nil
}
