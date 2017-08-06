package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	// tmpl := template.Must(template.ParseFiles("template.html"))
	// dir, _ := os.Getwd()
	// tmpl := template.Must(template.ParseFiles(filepath.Join(dir, "templates", "template.html")))
	// // テンプレートからテキストを生成して, os.Stdoutへ出力
	// err := tmpl.Execute(os.Stdout, nil)
	// if err != nil {
	// 	panic(err)
	// }

	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles("./src/helloworld/base.html"))
	// dat := struct {
	// 	Title string
	// 	Body  string
	// }{
	// 	Title: r.FormValue("title"),
	// 	Body:  r.FormValue("body"),
	// }
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

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "%v\n", handler.Header)

	f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	birthmark := r.FormValue("birthmark")
	fmt.Fprintf(w, "%s\n", birthmark)
	threshold := r.FormValue("threshold")
	fmt.Fprintf(w, "%s\n", threshold)

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
	fmt.Fprintf(w, "%s\n", resp)
	defer resp.Body.Close()

	// values := url.Values{}
	// values.Add("birthmark", birthmark)
	// values.Add("threshold", threshold)
	// values.Set("file", file)
	//
	// req, err := http.NewRequest(
	// 	"POST",
	// 	"http://localhost:9001/upload",
	// 	strings.NewReader(values.Encode()),
	// )
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	//
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// }
	// defer resp.Body.Close()

}

func main() {
	goji.Get("/", index)
	goji.Get("/hello/:name", hello)
	goji.Post("/hello", hello_post)
	goji.Post("/file", file)
	goji.Serve()
}
