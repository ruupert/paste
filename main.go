package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	pastedb "github.com/ruupert/paste/db"
	bolt "github.com/ruupert/paste/db/bolt"
)

type BodyData struct {
	Value string
}

var (
	goPastePort   int
	goPasteAddr   string
	goPasteTlsCrt string
	goPasteTlsKey string
	//backend       string
)

func init() {
	flag.IntVar(&goPastePort, "port", 8432, "Listen port")
	flag.StringVar(&goPasteAddr, "addr", "0.0.0.0", "Bind address")
	flag.StringVar(&goPasteTlsCrt, "cert", "tls.pem", "Cert")
	flag.StringVar(&goPasteTlsKey, "key", "tls.key", "Cert Key")
	//flag.StringVar(&backend, "db", "bolt", "backend options: bolt")
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		postHandler(w, r)
	case "GET":
		getHandler(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
	}
	var p pastedb.PasteRecord
	p.New(string(body))
	fmt.Println(p.Hash)
	bolt.Put(p)
	http.Redirect(w, r, "/"+p.Hash, http.StatusFound)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		getIndexHandler(w, r)
	} else {
		req, found := strings.CutPrefix(r.URL.Path, "/")
		if !found {
			http.Error(w, "ErrorPrefixNotFound", http.StatusInternalServerError)
		}
		res, err := bolt.Get(req)
		if err != nil {
			http.Error(w, "ErrorHashNotFound", http.StatusInternalServerError)
		}
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Cache-Control", "no-cache")
		pagedata := BodyData{Value: "<pre><code>" + res + "</code></pre>"}
		tmpl := template.Must(template.ParseFiles(wd + "/templates/layout.html"))
		tmpl.Execute(w, pagedata)
	}
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Cache-Control", "no-cache")
	pagedata := BodyData{Value: "<textarea class='prettyprint' id='paste' placeholder='[ paste text  -  ctrl+s to save ]' spellcheck='false'></textarea>"}
	tmpl := template.Must(template.ParseFiles(wd + "/templates/layout.html"))
	tmpl.Execute(w, pagedata)
}

func certsExist() bool {
	var cert bool = true
	var key bool = true
	if _, err := os.Stat(goPasteTlsCrt); errors.Is(err, os.ErrNotExist) {
		cert = false
	}
	if _, err := os.Stat(goPasteTlsKey); errors.Is(err, os.ErrNotExist) {
		key = false
	}
	if key && cert {
		return true
	} else {
		return false
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(requestHandler))
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	if certsExist() {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		srv := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", goPasteAddr, goPastePort),
			Handler:      mux,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		fmt.Printf("Listening https on %d\n", goPastePort)
		log.Fatal(srv.ListenAndServeTLS(goPasteTlsCrt, goPasteTlsKey))
	} else {
		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", goPasteAddr, goPastePort),
			Handler: mux,
		}
		fmt.Printf("Listening http on %d\n", goPastePort)
		log.Fatal(srv.ListenAndServe())
	}
}
