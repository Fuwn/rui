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
	Notify   bool   `json:"notify"`
	Editor   string `json:"editor"`
	Flake    string `json:"flake"`
	Notifier string `json:"notifier"`
}

type ActionDetails struct {
	Name         string
	Verb         string
	UsableWithNH bool
}

var configuration Configuration

const (
	Switch = iota
	Boot
	Test
	Build
	DryActivate
	BuildVM
	Instantiate
	Generations
	Packages
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

func subcommand(action int, aliases []string, flags []cli.Flag, commandAction func(c *cli.Context, action int) error) *cli.Command {
	return &cli.Command{
		Name:    actionName(action),
		Aliases: aliases,
		Flags:   flags,
		Action: func(c *cli.Context) error {
			return commandAction(c, action)
		},
	}
}

func main() {
	homeFlags := []cli.Flag{
		&cli.BoolFlag{
			Name: "use-home-manager",
		},
		&cli.StringFlag{
			Name: "user",
		},
	}
	osFlags := []cli.Flag{
		&cli.BoolFlag{
			Name: "use-nixos-rebuild",
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
					subcommand(Switch, []string{"sw"}, homeFlags, home),
					subcommand(Build, []string{}, homeFlags, home),
					subcommand(Instantiate, []string{}, homeFlags, home),
					subcommand(Generations, []string{"gens"}, homeFlags, home),
					subcommand(Packages, []string{"pkgs"}, homeFlags, home),
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
					subcommand(Switch, []string{"sw"}, osFlags, ruiOS),
					subcommand(Boot, []string{}, osFlags, ruiOS),
					subcommand(Test, []string{}, osFlags, ruiOS),
					subcommand(Build, []string{}, osFlags, ruiOS),
					subcommand(DryActivate, []string{"dry"}, osFlags, ruiOS),
					subcommand(BuildVM, []string{"vm"}, osFlags, ruiOS),
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

	notifier := configuration.Notifier

	if notifier == "" {
		notifier = "notify-send"
	}

	notifySend, err := exec.LookPath(notifier)

	if err != nil {
		return nil
	}

	if configuration.Notify {
		return command(notifySend, "Rui", message)
	}

	return nil
}

func actionName(action int) string {
	name := "switch"

	switch action {
	case Boot:
		name = "boot"

		break

	case Test:
		name = "test"

		break

	case Build:
		name = "build"

		break

	case DryActivate:
		name = "dry-activate"

		break

	case BuildVM:
		name = "build-vm"

		break

	case Instantiate:
		name = "instantiate"

		break

	case Generations:
		name = "generations"

		break

	case Packages:
		name = "packages"

		break
	}

	return name
}

func actionDetails(action int) (string, string, bool) {
	switch action {
	case Switch:
		return actionName(action), "switched", true

	case Boot:
		return actionName(action), "booted", false

	case Test:
		return actionName(action), "tested", false

	case Build:
		return actionName(action), "built", true

	case DryActivate:
		return actionName(action), "dry activated", false

	case BuildVM:
		return actionName(action), "VM built", false

	case Instantiate:
		return actionName(action), "instantiated", false

	case Generations:
		return actionName(action), "generations listed", false

	case Packages:
		return actionName(action), "packages shown", false
	}

	return "", "", false
}

func home(c *cli.Context, action int) error {
	nh, err := exec.LookPath("nh")
	extraArgs := c.Args().Slice()
	name, verb, usableWithNH := actionDetails(action)

	if err := notify("Queued home " + name); err != nil {
		return err
	}

	if err == nil && !c.Bool("use-home-manager") {
		if !usableWithNH {
			return fmt.Errorf("This command is not supported with nh. Use --use-home-manager to use Home Manager instead.")
		}

		err = command(nh, append([]string{"home", name, "--"},
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

		err = command("home-manager", append([]string{name,
			"--flake", fmt.Sprintf("%s#%s", flake, user)},
			extraArgs...)...)
	}

	if err != nil {
		return notify(fmt.Sprintf("Failed to %s home: %s", name, err.Error()))
	}

	return notify("Home " + verb)
}

func ruiOS(c *cli.Context, action int) error {
	nh, err := exec.LookPath("nh")
	name, verb, usableWithNH := actionDetails(action)

	if err := notify("Queued OS " + name); err != nil {
		return err
	}

	if err == nil && !c.Bool("use-nixos-rebuild") {
		if !usableWithNH {
			return fmt.Errorf("This command is not supported with nh. Use --use-nixos-rebuild to use nixos-rebuild instead.")
		}

		err = command(nh, "os", name)
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

		err = command(escalator, "nixos-rebuild", name, "--flake",
			fmt.Sprintf("%s#%s", flake, hostname))
	}

	if err != nil {
		return notify(fmt.Sprintf("Failed to %s OS: %s", name, err.Error()))
	}

	return notify("OS " + verb)
}
