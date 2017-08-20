package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type SearchResult struct {
	artifactId      string
	sim             string
	resultClassFile string
	groupId         string
	version         string
	postedClassFile string
	jar             string
	birthmarkData   string
}

type Config struct {
	Server []ServerConfig
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {

	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles("./base.html"))

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

	// get config
	var config Config
	_, err := toml.DecodeFile("./conf/config.tml", &config)
	if err != nil {
		panic(err)
	}

	// json格納用配列
	jsonStrings := []string{}

	for k, v := range config.Server {
		fmt.Printf("Slave %d\n", k)
		fmt.Printf("  weight is %d\n", v.Host)
		fmt.Printf("  ip is %s\n", v.Port)

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
		fmt.Println(birthmark)
		threshold := r.FormValue("threshold")
		fmt.Println(threshold)

		// request
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", filepath.Base("./test/"+handler.Filename))

		url := "http://" + v.Host + ":" + v.Port + "/upload"
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
			// fmt.Fprintf(w, "%s\n", string(b))
		}

		// json get
		var searchResult interface{}
		errr := json.Unmarshal([]byte(string(b)), &searchResult)
		if errr != nil {
			fmt.Println("error")
		}
		array := searchResult.([]interface{})

		for i := 0; i < len(array); i++ {
			json_str, _ := json.Marshal(array[i].(map[string]interface{}))
			jsonStrings = append(jsonStrings, string(json_str))
			fmt.Println(jsonStrings)
		}
	}
	fmt.Fprintf(w, "[%s]\n", strings.Join(jsonStrings[:], ","))
}

func download(c web.C, w http.ResponseWriter, r *http.Request) {
	searchResult := r.FormValue("searchResult")
	// fmt.Fprintf(w, "%s\n", searchResult)
	out := []byte(searchResult)
	// ファイル名
	w.Header().Set("Content-Disposition", "attachment; filename=result.json")
	// コンテントタイプ
	w.Header().Set("Content-Type", "application/json")
	// ファイルの長さ
	w.Header().Set("Content-Length", string(len(out)))
	// bodyに書き込み
	w.Write(out)
}

// compare
func compare(c web.C, w http.ResponseWriter, r *http.Request) {
	// get config
	var config Config
	_, err := toml.DecodeFile("./conf/config.tml", &config)
	if err != nil {
		panic(err)
	}

	jsonStrings := []string{}

	for k, v := range config.Server {
		fmt.Printf("Slave %d\n", k)
		searchResultForm := r.FormValue("searchResult")
		birthmark := r.FormValue("birthmark")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		url := "http://" + v.Host + ":" + v.Port + "/compare"
		params := map[string]string{
			"searchResult": searchResultForm,
			"birthmark":    birthmark,
		}
		for key, val := range params {
			_ = writer.WriteField(key, val)
		}

		err = writer.Close()
		req, err := http.NewRequest("POST", url, body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		client := &http.Client{}
		resp, err := client.Do(req)
		defer resp.Body.Close()

		fmt.Println("before response")

		// respose write
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			fmt.Println("b")
			fmt.Println(string(b))
			// fmt.Fprintf(w, "%s\n", string(b))
		}

		// json get
		var searchResult interface{}
		fmt.Println("Json get")
		fmt.Println([]byte(string(b)))
		errr := json.Unmarshal([]byte(string(b)), &searchResult)
		if errr != nil {
			fmt.Println("error")
		}
		array := searchResult.([]interface{})

		for i := 0; i < len(array); i++ {
			json_str, _ := json.Marshal(array[i].(map[string]interface{}))
			jsonStrings = append(jsonStrings, string(json_str))
			fmt.Println(jsonStrings)
		}
		fmt.Fprintf(w, "[%s]\n", strings.Join(jsonStrings[:], ","))

		break

	}
}

func main() {
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	goji.Get("/", index)
	goji.Get("/hello/:name", hello)
	goji.Post("/hello", hello_post)
	goji.Post("/file", file)
	goji.Post("/download", download)
	goji.Post("/compare", compare)
	goji.Serve()
}
