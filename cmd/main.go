package main

import (
	"log"
	"os"

	"github.com/cyrushanlon/music-organiser/organise"
	"github.com/cyrushanlon/music-organiser/sync"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:  "organise",
				Usage: "organise music library",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dummy", Aliases: []string{"d"}},
					&cli.BoolFlag{Name: "hideErrors", Aliases: []string{"e"}},
					&cli.StringFlag{
						Name:     "origin",
						Aliases:  []string{"o"},
						Required: true,
						Usage:    "path to the source files",
					},
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"t"},
						Required: true,
						Usage:    "path to where the organised file structure should be created",
					},
				},
				Action: func(c *cli.Context) error {
					return organise.Do(
						c.String("origin"),
						c.String("target"),
						c.Bool("hideErrors"),
						c.Bool("dummy"),
					)
				},
			},
			{
				Name:  "sync",
				Usage: "sync music",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "dummy", Aliases: []string{"d"}},
					&cli.StringFlag{
						Name:     "origin",
						Aliases:  []string{"o"},
						Required: true,
						Usage:    "path to the origin files",
					},
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"t"},
						Required: true,
						Usage:    "path to the sync target",
					},
				},
				Action: func(c *cli.Context) error {
					return sync.Do(c.String("origin"), c.String("target"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
