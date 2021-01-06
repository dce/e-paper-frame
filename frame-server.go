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
    <h1>My Photos</h1>

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
    http.Handle("/thumbs/",
        http.StripPrefix("/thumbs/", http.FileServer(http.Dir("./thumbs"))))

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

    _, err := os.Stat(fmt.Sprintf("thumbs/%s", photo))
    if err != nil {
        err = GenerateThumbnail(photo)

        if err != nil {
          fmt.Fprintf(w, "Error: %s", err)
          return
        }
    }

    http.Redirect(w, r, fmt.Sprintf("/thumbs/%s", photo), 301)
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

func GenerateThumbnail(filename string) (error) {
    file, err := os.Open(fmt.Sprintf("photos/%s", filename))
    if err != nil {
        return err
    }
    defer file.Close()

    img, err := jpeg.Decode(file)
    if err != nil {
        return err
    }

    thumb := resize.Resize(300, 0, img, resize.NearestNeighbor)

    err = os.MkdirAll("thumbs", 0755)
    if err != nil {
        return err
    }

    out, err := os.Create(fmt.Sprintf("thumbs/%s", filename))
    if err != nil {
        return err
    }
    defer out.Close()

    jpeg.Encode(out, thumb, nil)
    return nil
}
