package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "html/template"
)

var homepage = `<html>
  <body>
    <h1>My Bullshit</h1>

    <ul>
      {{range .}}
        <li><a href="/photos/{{.Name}}">{{.Name}}</a></li>
      {{end}}
  </body>
</html>`

func main() {
    http.Handle("/photos/",
        http.StripPrefix("/photos/", http.FileServer(http.Dir("./photos"))))
    http.HandleFunc("/", HelloServer)
    http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
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
