package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type PostData struct {
	birthmark string `json:"birthmark"`
	threshold string `json:"threshold"`
	file      byte   `json:"file"`
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles("./base.html"))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	err := templates.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func hello_post(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.FormValue("message"))
}

// file upload
func file(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	// file copy
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	// text value get
	birthmark := r.FormValue("birthmark")
	fmt.Printf(birthmark)
	threshold := r.FormValue("threshold")
	fmt.Printf(threshold)

	// request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base("./test/"+handler.Filename))

	url := "http://localhost:9000/upload"
	params := map[string]string{
		"birthmark": birthmark,
		"threshold": threshold,
	}

	_, err = io.Copy(part, file)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	// respose write
	b, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		fmt.Println(string(b))
		fmt.Fprintf(w, "%s\n", string(b))
	}
}

func main() {
	goji.Get("/", index)
	goji.Get("/hello/:name", hello)
	goji.Post("/hello", hello_post)
	goji.Post("/file", file)
	goji.Serve()
}
