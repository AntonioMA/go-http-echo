package main

import (
	"encoding/json"
	"flag"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	template2 "github.com/AntonioMA/go-http-echo/template"
	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/websocket"
)

type genericTemplate interface {
	Execute(wr io.Writer, data interface{}) error
}

func mirrorWebsocket(conn *websocket.Conn, req *http.Request) {
	// The only thing this websocket does is mirroring the input. Have fun.
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("read error: %v\n", err)
			break
		}
		fmt.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			fmt.Printf("write error: %v\n", err)
			break
		}
	}
}

func echoAll(outputTmpl genericTemplate, contentType string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("Processing request %v\n", req)

		if strings.ToLower(req.Header.Get("connection")) == "upgrade" && strings.ToLower(req.Header.Get("upgrade")) == "websocket" {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(res, req, nil)
			if err != nil {
				fmt.Printf("Error upgrading to websocket: %v\n", err)
				return
			}
			mirrorWebsocket(conn, req)
			return
		}

		res.Header().Set("Content-Type", contentType)
		res.WriteHeader(200)
		bodyAsStr := template2.ExtendedString("")
		if req.Body != nil {
			defer req.Body.Close()
			size := int(math.Min(float64(16384), float64(req.ContentLength)))
			if size <= 0 {
				size = 2048
			}
			buff := make([]byte, size)
			if n, err := req.Body.Read(buff); err != nil && err != io.EOF {
				bodyAsStr = template2.ExtendedString(fmt.Sprintf("Error reading body: %v", err))
			} else {
				bodyAsStr = template2.ExtendedString(buff[:n])
			}
		}
		tmplData := map[string]interface{}{
			"req":  req,
			"body": bodyAsStr,
		}
		_ = req.ParseForm()
		if req.Form.Get("json") == "1" || req.Header.Get("Accept") == "application/json" {
			tmplData["req"] = (*marshableRequest)(req)
			data, err := json.Marshal(tmplData)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				_, _ = res.Write(([]byte)(fmt.Sprintf(`{"error": "%s"}`, err)))
				return
			}
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusOK)
			_, _ = res.Write(data)
			return
		}

		if err := outputTmpl.Execute(res, tmplData); err != nil {
			_, _ = res.Write(([]byte)(fmt.Sprintf("<h3>Error executing template: %v</h3>", err)))
		}
		_ = outputTmpl.Execute(os.Stdout, tmplData)
	}
}

type marshableRequest http.Request

type auxRequest struct {
	Method           string
	URL              *url.URL
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	Pattern          string
}

func (r *marshableRequest) MarshalJSON() ([]byte, error) {
	aux := auxRequest{
		Method:           r.Method,
		URL:              r.URL,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Close:            r.Close,
		Host:             r.Host,
		Form:             r.Form,
		PostForm:         r.PostForm,
		MultipartForm:    r.MultipartForm,
		Trailer:          r.Trailer,
		RemoteAddr:       r.RemoteAddr,
		RequestURI:       r.RequestURI,
		Pattern:          r.Pattern,
	}
	return json.Marshal(aux)
}

func main() {
	var templatePath string
	var contentType string
	var debug bool
	flag.StringVar(&templatePath, "t", "./default_html.tmpl", "template path")
	flag.StringVar(&templatePath, "template", "./default_html.tmpl", "template path")
	flag.StringVar(&contentType, "c", "text/html", "content type")
	flag.StringVar(&contentType, "content-type", "text/html", "content type")
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.Parse()
	fileData, err := os.ReadFile(templatePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
		os.Exit(1)
	}
	if debug {
		fmt.Printf("Template from %s successfully read\nContents:\n%s\n", templatePath, string(fileData))
	}

	var tmpl genericTemplate
	if contentType != "text/html" {
		tmpl, err = template.New("dump").Funcs(sprig.FuncMap()).Parse(string(fileData))
	} else {
		tmpl, err = htmlTemplate.New("dump").Funcs(sprig.FuncMap()).Parse(string(fileData))
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}
	if err := http.ListenAndServe(":8128", http.HandlerFunc(echoAll(tmpl, contentType))); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
