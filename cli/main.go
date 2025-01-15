package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	var language string

	cmd := &cli.Command{
		Name:  "hello-cli",
		Usage: "Example command line binary that says hello",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Value:       "english",
				Usage:       "language for the greeting. will emit english if you give an unknown language. options are: spanish",
				Destination: &language,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var greeting string = "hello"
			if language == "spanish" {
				greeting = "hola"
			}
			fmt.Printf("%s %s, you have ran the example cli\n", greeting, cmd.Args().Get(0))
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
