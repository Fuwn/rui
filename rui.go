package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/urfave/cli/v2"
)

type Configuration struct {
	Notify bool   `json:"notify"`
	Editor string `json:"editor"`
	Flake  string `json:"flake"`
}

var configuration Configuration

const (
	Build = iota
	Switch
)

func init() {
	configurationPath := os.Getenv("RUI_CONFIG")

	if configurationPath == "" {
		configurationPath = os.Getenv("XDG_CONFIG_HOME") + "/rui/config.json"
	}

	content, err := os.ReadFile(configurationPath)

	if err != nil {
		return
	}

	if err := json.Unmarshal(content, &configuration); err != nil {
		return
	}
}

func main() {
	homeFlags := []cli.Flag{
		&cli.BoolFlag{
			Name: "force-home-manager",
		},
		&cli.StringFlag{
			Name: "user",
		},
	}
	osFlags := []cli.Flag{
		&cli.BoolFlag{
			Name: "force-nixos-rebuild",
		},
		&cli.StringFlag{
			Name: "hostname",
		},
	}

	(&cli.App{
		Name:                 "rui",
		Usage:                "Personal NixOS Flake Manager",
		Description:          "Personal NixOS Flake Manager",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Fuwn",
				Email: "contact@fuwn.me",
			},
		},
		Copyright: fmt.Sprintf("Copyright (c) 2024-%s Fuwn", fmt.Sprint(time.Now().Year())),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Println(err)
			}
		},
		Suggest: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "allow-unfree",
				Action: func(c *cli.Context, b bool) error {
					state := "0"

					if b {
						state = "1"
					}

					return os.Setenv("NIXPKGS_ALLOW_UNFREE", state)
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name: "hs",
				Action: func(c *cli.Context) error {
					return c.App.Command("home").Command("switch").Run(c)
				},
				Hidden:      true,
				Description: "Alias for `home switch`",
			},
			{
				Name: "osw",
				Action: func(c *cli.Context) error {
					return c.App.Command("os").Command("switch").Run(c)
				},
				Hidden: true,
				Usage:  "Alias for `os switch`",
			},
			{
				Name: "home",
				Subcommands: []*cli.Command{
					{
						Name:    "switch",
						Aliases: []string{"sw"},
						Flags:   homeFlags,
						Action: func(c *cli.Context) error {
							return home(c, Switch)
						},
					},
					{
						Name:  "build",
						Flags: homeFlags,
						Action: func(c *cli.Context) error {
							return home(c, Build)
						},
					},
					{
						Name: "news",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "user",
							},
						},
						Action: func(c *cli.Context) error {
							flake := configuration.Flake
							extraArgs := c.Args().Slice()

							if flake == "" {
								flake = os.Getenv("FLAKE")
							}

							if user := c.String("user"); user != "" {
								flake = fmt.Sprintf("%s#%s", flake, user)
							}

							return command("home-manager", append([]string{"news", "--flake",
								flake}, extraArgs...)...)
						},
					},
				},
			},
			{
				Name: "os",
				Subcommands: []*cli.Command{
					{
						Name:    "switch",
						Aliases: []string{"sw"},
						Flags:   osFlags,
						Action: func(c *cli.Context) error {
							return ruiOS(c, Switch)
						},
					},
					{
						Name:  "build",
						Flags: osFlags,
						Action: func(c *cli.Context) error {
							return ruiOS(c, Build)
						},
					},
				},
			},
			{
				Name: "edit",
				Action: func(c *cli.Context) error {
					var found bool

					editor := configuration.Editor
					flake := configuration.Flake

					if flake == "" {
						flake = os.Getenv("FLAKE")
					}

					if editor == "" {
						if editor, found = os.LookupEnv("FLAKE_EDITOR"); !found {
							editor = os.Getenv("EDITOR")
						}
					}

					return command(editor, flake)
				},
			},
		},
	}).Run(os.Args)
}

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func notify(message string) error {
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return nil
	}

	notifySend, err := exec.LookPath("notify-send")

	if err != nil {
		return nil
	}

	if configuration.Notify {
		return command(notifySend, "Rui", message)
	}

	return nil
}

func actionName(action int) string {
	if action == Build {
		return "build"
	}

	return "switch"
}

func actionVerb(action int) string {
	if action == Build {
		return "built"
	}

	return "switched"
}

func home(c *cli.Context, action int) error {
	nh, err := exec.LookPath("nh")
	extraArgs := c.Args().Slice()
	actionName := actionName(action)

	if err := notify("Queued home " + actionName); err != nil {
		return err
	}

	if err == nil && !c.Bool("force-home-manager") {
		err = command(nh, append([]string{"home", actionName, "--"},
			extraArgs...)...)
	} else {
		user := c.String("user")

		if user == "" {
			user = os.Getenv("USER")
		}

		flake := configuration.Flake

		if flake == "" {
			flake = os.Getenv("FLAKE")
		}

		err = command("home-manager", append([]string{actionName,
			"--flake", fmt.Sprintf("%s#%s", flake, user)},
			extraArgs...)...)
	}

	if err != nil {
		return notify(fmt.Sprintf("Failed to %s home: %s", actionName, err.Error()))
	}

	return notify("Home " + actionVerb(action))
}

func ruiOS(c *cli.Context, action int) error {
	nh, err := exec.LookPath("nh")
	actionName := actionName(action)

	if err := notify("Queued OS " + actionName); err != nil {
		return err
	}

	if err == nil && !c.Bool("force-nixos-rebuild") {
		err = command(nh, "os", actionName)
	} else {
		escalator := "sudo"

		if doas, err := exec.LookPath("doas"); err != nil {
			escalator = doas
		}

		hostname := c.String("hostname")

		if hostname == "" {
			hostname, err = os.Hostname()

			if err != nil {
				return err
			}
		}

		flake := configuration.Flake

		if flake == "" {
			flake = os.Getenv("FLAKE")
		}

		err = command(escalator, "nixos-rebuild", actionName, "--flake",
			fmt.Sprintf("%s#%s", flake, hostname))
	}

	if err != nil {
		return notify(fmt.Sprintf("Failed to %s OS: %s", actionName, err.Error()))
	}

	return notify("OS " + actionVerb(action))
}
