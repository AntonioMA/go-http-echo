package main

import (
	"flag"
	"fmt"
	template2 "github.com/AntonioMA/go-http-echo/template"
	"github.com/masterminds/sprig"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
)

func echoAll(outputTmpl *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("Processing request %v\n", req)
		res.Header().Set("Content-Type", "text/html")
		res.WriteHeader(200)
		bodyAsStr := template2.ExtendedString("")
		if req.Body != nil {
			defer req.Body.Close()
			size := int(math.Min(float64(2048), float64(req.ContentLength)))
			if size <= 0 { // Gotta love Armadillo
				size = 2048
			}
			buff := make([]byte, size)
			if n, err := req.Body.Read(buff); err != nil && err != io.EOF {
				bodyAsStr = template2.ExtendedString(fmt.Sprintf("Error reading body: %v", err))
			} else {
				bodyAsStr = template2.ExtendedString(buff[:n])
				bodyAsStr = bodyAsStr + "\n" + bodyAsStr.ToBase64()
			}
		}
		if err := outputTmpl.Execute(res, map[string]interface{}{
			"req":  req,
			"body": bodyAsStr,
		}); err != nil {
			_, _ = res.Write(([]byte)(fmt.Sprintf("<h3>Error executing template: %v</h3>", err)))
		}
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

	tmpl, err := template.New("dump").Funcs(sprig.FuncMap()).Parse(string(fileData))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}
	if err := http.ListenAndServe(":8128", http.HandlerFunc(echoAll(tmpl))); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
