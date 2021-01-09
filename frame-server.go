package main

import (
	"fmt"
	"github.com/dce/rpi/epd7in5"
	"html/template"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var homepage = `<!DOCTYPE html>
	<html>
		<body>
			{{range .}}
				<p>
					<a href="/photos/{{.Name}}">
						<img src="/thumb?p={{.Name}}">
					</a><br>
					<a href="/dither?p={{.Name}}">Dither</a>
					<a href="/display?p={{.Name}}">Display</a>
				</p>
		{{end}}
	</body>
</html>`

func main() {
	http.Handle("/photos/",
		http.StripPrefix("/photos/", http.FileServer(http.Dir("./photos"))))
	http.Handle("/thumbs/",
		http.StripPrefix("/thumbs/", http.FileServer(http.Dir("./thumbs"))))
	http.Handle("/dithered/",
		http.StripPrefix("/dithered/", http.FileServer(http.Dir("./dithered"))))

	http.HandleFunc("/thumb", Thumbnail)
	http.HandleFunc("/dither", Dither)
	http.HandleFunc("/display", Display)
	http.HandleFunc("/", ListPhotos)

	http.ListenAndServe(":8080", nil)
}

func Thumbnail(w http.ResponseWriter, r *http.Request) {
	photos, ok := r.URL.Query()["p"]
	if !ok {
		fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
		return
	}

	photo := photos[0]
	thumbPath := path("thumbs", photo)

	_, err := os.Stat(thumbPath)
	if err != nil {
		err = GenerateThumbnail(photo)

		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", thumbPath), 301)
}

func Dither(w http.ResponseWriter, r *http.Request) {
	photos, ok := r.URL.Query()["p"]
	if !ok {
		fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
		return
	}

	photo := photos[0]
	dithered := path("dithered", photo)

	_, err := os.Stat(dithered)
	if err != nil {
		err = GenerateDitheredImage(photo)

		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", dithered), 301)
}

func Display(w http.ResponseWriter, r *http.Request) {
	photos, ok := r.URL.Query()["p"]
	if !ok {
		fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
		return
	}

	photo := photos[0]

	log.Println("Starting...")
	epd, _ := epd7in5.New("P1_22", "P1_24", "P1_11", "P1_18")

	file, err := os.Open(path("dithered", photo))
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.Println("Initializing the display...")
	epd.Init()

	log.Println("Clearing...")
	epd.Clear()

	log.Println("Displaying image...")
	epd.Display(epd.Convert(img))

	fmt.Fprintf(w, "Displaying %s", photo)
}

func ListPhotos(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./photos")
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	tmpl := template.New("Page")
	if _, err := tmpl.Parse(homepage); err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	tmpl.Execute(w, files)
}

func GenerateThumbnail(filename string) error {
	err := os.MkdirAll("thumbs", 0755)
	if err != nil {
		return err
	}

	err = convert(
		path("photos", filename),
		path("thumbs", filename),
		"-resize",
		"400x300",
		"-gravity",
		"center",
		"-extent",
		"400x300",
	)

	if err != nil {
		return err
	}
	return nil
}

func GenerateDitheredImage(filename string) error {
	err := os.MkdirAll("dithered", 0755)
	if err != nil {
		return err
	}

	err = convert(
		path("photos", filename),
		path("dithered", filename),
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
