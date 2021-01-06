package main

import (
    "fmt"
    "html/template"
    "image/jpeg"
    "io/ioutil"
    "github.com/nfnt/resize"
    "net/http"
    "os"
)

var homepage = `<!DOCTYPE html>
<html>
  <body>
    <h1>My Bullshit</h1>

    <ul>
      {{range .}}
        <li>
          <a href="/photos/{{.Name}}">
            <img src="/thumb?p={{.Name}}">
            {{.Name}}
          </a>
        </li>
      {{end}}
    </ul>
  </body>
</html>`

func main() {
    http.Handle("/photos/",
        http.StripPrefix("/photos/", http.FileServer(http.Dir("./photos"))))
    http.HandleFunc("/thumb", Thumbnail)
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

    file, err := os.Open(fmt.Sprintf("./photos/%s", photo))
    if err != nil {
        fmt.Fprintf(w, "Error: file %s not found", photo)
        return
    }

    img, err := jpeg.Decode(file)
    if err != nil {
        fmt.Fprintf(w, "Error: file %s could not be decoded", photo)
        return
    }

    thumb := resize.Resize(300, 0, img, resize.NearestNeighbor)
    jpeg.Encode(w, thumb, nil)
}

func ListPhotos(w http.ResponseWriter, r *http.Request) {
    files, err := ioutil.ReadDir("./photos")

    if err != nil {
        fmt.Fprintf(w, "Error: %s", err)
    } else {
        tmpl := template.New("Page")

        if tmpl, err := tmpl.Parse(homepage); err != nil {
            fmt.Fprintf(w, "Error: %s", err)
        } else {
            tmpl.Execute(w, files)
        }
    }
}
