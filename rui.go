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
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name: "force-home-manager",
							},
							&cli.StringFlag{
								Name: "user",
							},
						},
						Action: func(c *cli.Context) error {
							nh, err := exec.LookPath("nh")
							extraArgs := c.Args().Slice()

							if err := notify("Queued home switch"); err != nil {
								return err
							}

							if err == nil && !c.Bool("force-home-manager") {
								err = command(nh, append([]string{"home", "switch", "--"},
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

								err = command("home-manager", append([]string{"switch",
									"--flake", fmt.Sprintf("%s#%s", flake, user)},
									extraArgs...)...)
							}

							if err != nil {
								return notify(fmt.Sprintf("Failed to switch home: %s", err.Error()))
							}

							return notify("Home switched")
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
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name: "force-nixos-rebuild",
							},
							&cli.StringFlag{
								Name: "hostname",
							},
						},
						Action: func(c *cli.Context) error {
							nh, err := exec.LookPath("nh")

							if err := notify("Queued OS switch"); err != nil {
								return err
							}

							if err == nil && !c.Bool("force-nixos-rebuild") {
								err = command(nh, "os", "switch")
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

								err = command(escalator, "nixos-rebuild", "switch", "--flake",
									fmt.Sprintf("%s#%s", flake, hostname))
							}

							if err != nil {
								return notify(fmt.Sprintf("Failed to switch OS: %s", err.Error()))
							}

							return notify("OS switched")
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
