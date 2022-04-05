package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"text/template"

	template2 "github.com/AntonioMA/go-http-echo/template"
	"github.com/gorilla/websocket"
	"github.com/masterminds/sprig"
)

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

func echoAll(outputTmpl *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("Processing request %v\n", req)

		if req.Header.Get("connection") == "Upgrade" && req.Header.Get("upgrade") == "websocket" {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(res, req, nil)
			if err != nil {
				fmt.Printf("Error upgrading to websocket: %v\n", err)
				return
			}
			mirrorWebsocket(conn, req)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.WriteHeader(200)
		bodyAsStr := template2.ExtendedString("")
		if req.Body != nil {
			defer req.Body.Close()
			size := int(math.Min(float64(16384), float64(req.ContentLength)))
			if size <= 0 { // Gotta love Armadillo
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
		if err := outputTmpl.Execute(res, tmplData); err != nil {
			_, _ = res.Write(([]byte)(fmt.Sprintf("<h3>Error executing template: %v</h3>", err)))
		}
		_ = outputTmpl.Execute(os.Stdout, tmplData)
	}
}

func main() {
	var templatePath string
	flag.StringVar(&templatePath, "t", "./default_html.tmpl", "template path")
	flag.StringVar(&templatePath, "template", "./default_html.tmpl", "template path")
	flag.Parse()
	fileData, err := os.ReadFile(templatePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Template from %s successfully read\nContents:\n%s\n", templatePath, string(fileData))

	tmpl, err := template.New("dump").Funcs(template.FuncMap(sprig.FuncMap())).Parse(string(fileData))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}
	if err := http.ListenAndServe(":8128", http.HandlerFunc(echoAll(tmpl))); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
