package main

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

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

//go:embed templates/dir.html
var dirTemplate string

//go:embed templates/file.html
var fileTemplate string

// define the dir struct
// define the file struct defining file dir, name and size

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

func getArgType(dir string) string {
	// check if the dir is a file or a directory
	inf, err := os.Stat(dir)

	if err != nil {
		log.Fatal(err)
	}
	if inf.IsDir() {
		return "dir"
	} else {
		return "file"
	}
}

func renderDirHTML(writer http.ResponseWriter, folder Dir, zipPath string, zipName string) {
	// render the directory template

	fld := Dir{
		DirName: folder.DirName,
		DirPath: folder.DirPath,
		Files:   folder.Files,
		ZipPath: zipPath,
		ZipName: zipName,
	}

	log.Println("Rendering ", fld.DirName, " at ", fld.DirPath)
	log.Println(fld)

	tmpl, err := template.New("dir").Parse(dirTemplate)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(writer, fld)
	if err != nil {
		log.Fatal(err)
	}

}

func renderFileHTML(writer http.ResponseWriter, file File) {

	f := File{
		Path: file.Path,
		Name: file.Name,
		Size: file.Size,
	}

	log.Println("Rendering ", f.Name, " at ", f.Path)
	log.Println(f)

	tmpl, err := template.New("file").Parse(fileTemplate)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(writer, f)
	if err != nil {
		log.Fatal(err)
	}

}

func run(c *cli.Context) error {
	fmt.Println("Boar is running...")
	arg := c.Args().Get(0)

	target, err := filepath.Abs(arg)
	if err != nil {
		log.Fatal(err)
	}

	if arg == "" {
		log.Fatal("No directory nor file provided.")
	}
	argtype := getArgType(target)
	log.Println("The argument is a ", argtype)
	if argtype == "dir" {
		zipPath := zipFolder(target)
		log.Println("The zip path is ", zipPath)

		dirToServe := Dir{
			DirName: filepath.Base(target),
			DirPath: target,
			ZipPath: zipPath,
			ZipName: filepath.Base(zipPath),
			Files:   retrieveFiles(target),
		}
		webServer(dirToServe, c.String("port"), argtype, zipPath, File{})

	} else if argtype == "file" {
		file := File{
			Path: target,
			Name: filepath.Base(target),
			Size: 0,
		}
		webServer(Dir{}, c.String("port"), argtype, "", file)
	}
	return nil
}

func retrieveFiles(dir string) []File {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var fls []File
	for _, file := range files {

		i, err := file.Info()
		if err != nil {
			log.Fatal(err)
		}

		f := File{
			Path: filepath.Join(dir, file.Name()),
			Name: file.Name(),
			Size: i.Size(),
		}
		fls = append(fls, f)
	}
	return fls
}

type WebHandler struct {
	targetType string
	dirToServe Dir
	zipPath    string
	zipName    string

	//File
	file File
}

func (wh *WebHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	t := wh.targetType
	zP := wh.zipPath
	dTS := wh.dirToServe
	zN := wh.zipName

	fF := wh.file

	if t == "dir" {
		renderDirHTML(w, dTS, zP, zN)
	} else {
		renderFileHTML(w, fF)
	}
}

func webServer(dirToServe Dir, port string, argtype string, zipPath string, file File) {
	log.Println("Serving ", dirToServe.DirName, " at ", dirToServe.DirPath)
	wh := WebHandler{
		targetType: argtype,
		dirToServe: dirToServe,
		zipPath:    zipPath,
		zipName:    filepath.Base(zipPath),
		file:       file,
	}
	http.Handle("/", http.HandlerFunc(wh.HandleRequest))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
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
