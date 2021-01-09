package main

import (
    "fmt"
    "html/template"
    "image"
    "image/jpeg"
    "io/ioutil"
    "github.com/nfnt/resize"
    "github.com/dce/rpi/epd7in5"
    "github.com/disintegration/gift"
    "github.com/lestrrat-go/dither"
    "log"
    "net/http"
    "os"
)

var homepage = `<!DOCTYPE html>
<html>
  <body>
    {{range .}}
      <p>
        <a href="/photos/{{.Name}}">
          <img src="/thumb?p={{.Name}}">
        </a><br>
        <a href="/thumb2?p={{.Name}}">Thumb</a>
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

    http.HandleFunc("/thumb", Thumbnail)
    http.HandleFunc("/thumb2", Thumbnail2)
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

func Thumbnail2(w http.ResponseWriter, r *http.Request) {
    photos, ok := r.URL.Query()["p"]
    if !ok {
        fmt.Fprintf(w, "Error: required parameter ('p') not supplied")
        return
    }

    photo := photos[0]

    file, err := os.Open(fmt.Sprintf("photos/%s", photo))
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

    g := gift.New(
      gift.ResizeToFill(880, 852, gift.LanczosResampling, gift.CenterAnchor),
    )

    thumb := image.NewRGBA(g.Bounds(img.Bounds()))

    g.Draw(thumb, img)

    dithered := dither.Monochrome(dither.Burkes, thumb, 1.18)

    jpeg.Encode(w, dithered, nil)
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

    file, err := os.Open(fmt.Sprintf("photos/%s", photo))
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
