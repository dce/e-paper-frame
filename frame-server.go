package main

import (
	"fmt"
	"github.com/dce/rpi/epd7in5"
	"html/template"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"os"
	"os/exec"
	"time"
)

type homepageData struct {
	Flash  string
	PhotoRows [][]os.FileInfo
}

var homepage = `<!DOCTYPE html>
<html>
	<head>
		<link
			href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta1/dist/css/bootstrap.min.css"
			rel="stylesheet"
			integrity="sha384-giJF6kkoqNQ00vy+HMDP7azOuL0xtbfIcaT9wjKHr8RbDVddVHyTfAAsrekwKmP1"
			crossorigin="anonymous">
	</head>

	<body>
		<div class="container mt-5">
			{{if .Flash}}
				<div class="alert alert-success" role="alert">
					{{.Flash}}
				</div>
			{{end}}

			{{range .PhotoRows}}
				<div class="row">
					{{range .}}
						<div class="col-md-4 text-center mb-5">
							<a href="/photos/{{.Name}}">
								<img src="/thumb?p={{.Name}}" class="img-fluid">
							</a><br>
							<a href="/display?p={{.Name}}" class="btn btn-primary mt-2">Display</a>
						</div>
					{{end}}
				</div>
			{{end}}
		</div>
	</body>
</html>`

func main() {
	if len(os.Args) == 1 {
		log.Println("Please supply a command")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "serve":
		startServer()
	case "random":
		displayRandomPhoto()
	default:
		log.Println("Unrecognized command:", cmd)
	}
}

func displayRandomPhoto() {
	photos, err := ioutil.ReadDir("./photos")
	if err != nil {
		log.Println("Error:", err)
		return
	}

	rand.Seed(time.Now().UnixNano())
	photo := photos[rand.Intn(len(photos))]
	displayPhoto(photo.Name())
}

func startServer() {
	log.Println("Server is starting")

	http.Handle("/photos/",
		http.StripPrefix("/photos/", http.FileServer(http.Dir("./photos"))))
	http.Handle("/thumbs/",
		http.StripPrefix("/thumbs/", http.FileServer(http.Dir("./thumbs"))))

	http.HandleFunc("/thumb", thumbHandler)
	http.HandleFunc("/display", displayHandler)
	http.HandleFunc("/", indexHandler)

	http.ListenAndServe(":80", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./photos")
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	var msg string
	if flash, err := r.Cookie("flash"); err == nil {
		msg = flash.Value
	} else {
		msg = ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "flash",
		Value:  "",
		MaxAge: -1,
	})

	tmpl := template.New("Page")
	if _, err := tmpl.Parse(homepage); err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	tmpl.Execute(w, &homepageData{
		Flash:  msg,
		PhotoRows: photoRows(files),
	})
}

func thumbHandler(w http.ResponseWriter, r *http.Request) {
	photos, ok := r.URL.Query()["p"]
	if !ok {
		fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
		return
	}

	photo := photos[0]
	thumbPath := path("thumbs", photo)

	if _, err := os.Stat(thumbPath); err != nil {
		if err = generateThumbnail(photo); err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", thumbPath), 301)
}

func displayHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Displaying photo")

	photos, ok := r.URL.Query()["p"]
	if !ok {
		fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
		return
	}

	photo := photos[0]

	if err := displayPhoto(photo); err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "flash",
		Value: fmt.Sprintf("Photo %s displayed!", photo),
	})
	http.Redirect(w, r, "/", 302)
}

func displayPhoto(filename string) error {
	dithered := path("dithered", filename)

	if _, err := os.Stat(dithered); err != nil {
		if err = generateDitheredImage(filename); err != nil {
			return err
		}
	}

	file, err := os.Open(dithered)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	if runtime.GOARCH != "arm" {
		return nil
	}

	epd, _ := epd7in5.New("P1_22", "P1_24", "P1_11", "P1_18")

	log.Println("-> Initializing the display")
	epd.Init()

	log.Println("-> Clearing")
	epd.Clear()

	log.Println("-> Displaying", filename)
	epd.Display(epd.Convert(img))

	return nil
}

func generateThumbnail(filename string) error {
	log.Println("-> Generating thumbnail for", filename)

	err := os.MkdirAll("thumbs", 0755)
	if err != nil {
		return err
	}

	err = convert(
		path("photos", filename),
		path("thumbs", filename),
		"-auto-orient",
		"-resize",
		"640x480^",
		"-gravity",
		"center",
		"-extent",
		"640x480",
	)

	if err != nil {
		return err
	}
	return nil
}

func generateDitheredImage(filename string) error {
	log.Println("-> Generating dithered image for", filename)

	err := os.MkdirAll("dithered", 0755)
	if err != nil {
		return err
	}

	err = convert(
		path("photos", filename),
		path("dithered", filename),
		"-auto-orient",
		"-resize",
		"880x528^",
		"-gravity",
		"center",
		"-extent",
		"880x528",
		"-monochrome",
		"-dither",
		"Riemersma",
		"-negate",
	)

	if err != nil {
		return err
	}

	return nil
}

func convert(infile string, outfile string, options ...string) error {
	var args []string
	args = append(args, infile)
	args = append(args, options...)
	args = append(args, outfile)

	cmd := exec.Command("convert", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func path(dir string, file string) string {
	return fmt.Sprintf("%s/%s", dir, file)
}

func photoRows(photos []os.FileInfo) ([][]os.FileInfo) {
	rows := make([][]os.FileInfo, 0)

	for i := 0; i < len(photos); i += 3 {
		if i + 2 < len(photos) {
			rows = append(rows, photos[i:i+3])
		} else {
			rows = append(rows, photos[i:])
		}
	}
	return rows
}
