package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)

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
        w.Header().Set("Content-Type", "text/html; charset=UTF-8")
        fmt.Fprintf(w, "<ul>")
        for _, file := range files {
            fmt.Fprintf(w, "<li><a href=\"/photos/%s\">%s</a></li>",
                file.Name(), file.Name())
        }
        fmt.Fprintf(w, "</ul>")
    }
}
