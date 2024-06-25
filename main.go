package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func getTempZipPath(parent string) string {
	// get the temp directory to store the zip file
	// create a random string to append to the parent directory
	salt := strconv.FormatInt(time.Now().UnixMilli(), 10)
	tempdir := filepath.Join(parent, salt+"boar.zip.tmp")
	return tempdir
}
func rmtempzipdir(dir string) {

	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
}

func zipFolder(folder string) string {
	log.Println("Zipping ", folder)
	// zip the directory
	zipdir := getTempZipPath(folder)

	zipFile, err := os.Create(zipdir)
	if err != nil {
		return err.Error()

	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		log.Println("Walking", path)

		if err != nil {
			log.Println("Error walking", path, ":", err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Println("Error creating file header for", path, ":", err)
			return err
		}

		// Set the name of the file within the zip archive
		fileHeader.Name, err = filepath.Rel(folder, path)
		if err != nil {
			log.Println("Error getting relative path for", path, ":", err)
			return err
		}

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			log.Println("Error opening file", path, ":", err)
			return err
		}
		defer file.Close()

		writer, err := archive.CreateHeader(fileHeader)
		if err != nil {
			log.Println("Error creating writer for", path, ":", err)
			return err
		}

		// Copy the file's contents to the zip archive
		_, err = io.Copy(writer, file)
		if err != nil {
			log.Println("Error copying file", path, ":", err)
			return err
		}

		return nil
	})

	log.Println("Finished walking")

	if err != nil {
		fmt.Println("Error zipping directory:", err)
		return err.Error()
	}

	fmt.Println("Directory zipped successfully.")
	return zipdir

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
