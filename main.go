package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

// Boar is a simple CLI that creates a web server that serves static files or directories. When a directory is specified, Boar will serve a list with the files to download each individually or as a zip file. When a file is specified, Boar will serve the file for download. at /<file> or a simple UI to download the file at /.

// Boar uses the following command line arguments:
// -p, --port: The port to listen on. Default is 8080.
// nz, --nozip: Disable the zip download feature. Default is false.
// -h, --help: Display the help message.
// -v, --version: Display the version number.
// -c, --children: Serve the children of the directory as ZIP. Default is false.

// Boar uses the following environment variables:
// BOAR_PORT: The port to listen on. Default is 8080.
// BOAR_NOZIP: Disable the zip download feature. Default is false.

type File struct {
	Path string
	Name string
	Size int64
}
type Dir struct {
	DirName string
	DirPath string
	Files   []File
	ZipPath string
	ZipName string
}

func run(c *cli.Context) error {
	port := c.Int("port")
	nozip := c.Bool("nozip")
	children := c.Bool("children")

	log.Printf("Starting Boar on port %d", port)

	return nil
}

func main() {
	app := &cli.App{
		Name:  "Boar",
		Usage: "A simple CLI to share files throught http(s).",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "nozip",
				Aliases: []string{"nz"},
				Value:   false,
				EnvVars: []string{"BOAR_NOZIP"},
				Usage:   "Disable the zip download feature.",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   8080,
				EnvVars: []string{"BOAR_PORT"},
				Usage:   "The port to listen on.",
			},
			&cli.BoolFlag{
				Name:    "children",
				Aliases: []string{"c"},
				Value:   false,
				EnvVars: []string{"BOAR_CHILDREN"},
				Usage:   "Serve the children of the directory as ZIP.",
			},
		},

		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
