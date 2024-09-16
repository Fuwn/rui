package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	(&cli.App{
		Name:                 "rui",
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
				Name: "home",
				Subcommands: []*cli.Command{
					{
						Name: "switch",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "impure",
								Value: true,
							},
							&cli.BoolFlag{
								Name: "force-home-manager",
							},
							&cli.StringFlag{
								Name: "user",
							},
						},
						Action: func(c *cli.Context) error {
							_, err := exec.LookPath("nh")
							extraArgs := []string{}

							if c.Bool("impure") {
								extraArgs = []string{"--impure"}
							}

							if err == nil && !c.Bool("force-home-manager") {
								return Command("nh", append([]string{"home", "switch", "--"},
									extraArgs...)...)
							}

							user := c.String("user")

							if user == "" {
								user = os.Getenv("USER")
							}

							return Command("home-manager", append([]string{"switch",
								"--flake", fmt.Sprintf("%s#%s", os.Getenv("FLAKE"), user)},
								extraArgs...)...)
						},
					},
					{
						Name: "news",
						Action: func(c *cli.Context) error {
							return Command("home-manager", "news", "--flake",
								os.Getenv("FLAKE"), "--impure")
						},
					},
				},
			},
			{
				Name: "os",
				Subcommands: []*cli.Command{
					{
						Name: "switch",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name: "force-nixos-rebuild",
							},
							&cli.StringFlag{
								Name: "hostname",
							},
						},
						Action: func(c *cli.Context) error {
							_, err := exec.LookPath("nh")

							if err == nil && !c.Bool("force-nixos-rebuild") {
								return Command("nh", "os", "switch")
							}

							_, err = exec.LookPath("doas")
							escalator := "sudo"

							if err == nil {
								escalator = "doas"
							}

							hostname := c.String("hostname")

							if hostname == "" {
								hostname, err = os.Hostname()

								if err != nil {
									return err
								}
							}

							return Command(escalator, "nixos-rebuild", "switch", "--flake",
								fmt.Sprintf("%s#%s", os.Getenv("FLAKE"), hostname))
						},
					},
				},
			},
			{
				Name: "edit",
				Action: func(c *cli.Context) error {
					editor, err := os.LookupEnv("FLAKE_EDITOR")

					if err {
						return Command(editor, os.Getenv("FLAKE"))
					}

					return Command(os.Getenv("EDITOR"), os.Getenv("FLAKE"))
				},
			},
		},
	}).Run(os.Args)
}

func Command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}