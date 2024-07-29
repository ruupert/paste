package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"

	"embed"

	"github.com/grafana/pyroscope-go"
	pastedb "github.com/ruupert/paste/backend/db"
)

type BodyData struct {
	Value string
}

var (
	db               pastedb.DatabaseInterface
	goPastePort      int
	goPasteAddr      string
	goPasteTlsCrt    string
	goPasteTlsKey    string
	goPasteDb        string
	goPaste404Dir    string
	goPastePyroscope string
)

//go:embed public/out.wasm
//go:embed public/script/wasm_exec.js
//go:embed public/css/paste.css
var embedded embed.FS

func init() {
	flag.IntVar(&goPastePort, "port", 8432, "Listen port")
	flag.StringVar(&goPasteAddr, "addr", "0.0.0.0", "Bind address")
	flag.StringVar(&goPasteTlsCrt, "cert", "tls.pem", "Cert")
	flag.StringVar(&goPasteTlsKey, "key", "tls.key", "Cert Key")
	flag.StringVar(&goPasteDb, "db", "bolt", "backend options: [bolt, memory]")
	flag.StringVar(&goPaste404Dir, "404", "./backend/public/404", "Path to dir with 404 png/gif/jpg") // later, just default for now
	flag.StringVar(&goPastePyroscope, "pyroscope", "", "Pyroscope ServerAddress")
	flag.Parse()
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	switch r.Method {
	case "POST":
		fmt.Println("Post")
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
	p.New(body)
	_, err = db.Put(p)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Add("Location", "/"+string(p.Hash))
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("ok"))
	if err != nil {
		fmt.Println(err)
	}
	//http.Redirect(w, r, "/"+string(p.Hash), http.StatusCreated)

}

func textOnly(a string) bool {
	s := strings.ToLower(a)
	r := []string{"curl", "wget", "fetch"}
	for _, v := range r {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		getIndexHandler(w)
		return
	}
	if r.URL.Path == "/wf" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Cache-Control", "no-cache")
		pagedata := BodyData{Value: ""}
		tmpl := template.Must(template.ParseFiles(wd + "/backend/templates/wlayout.html"))
		err = tmpl.Execute(w, pagedata)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if r.URL.Path == "/wasssm" {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFileFS(w, r, embedded, "public/out.wasm")
		return
	}
	if r.URL.Path == "/css" {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFileFS(w, r, embedded, "public/css/paste.css")
		return
	}
	if r.URL.Path == "/wasmexec" {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFileFS(w, r, embedded, "public/script/wasm_exec.js")
		return
	}
	req, found := strings.CutPrefix(r.URL.Path, "/")
	if !found {
		notFoundHandler(w)
		return
	}
	res, err := db.Get([]byte(req))
	if err != nil {
		notFoundHandler(w)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if textOnly(r.UserAgent()) {
		_, err = w.Write(res)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		w.Header().Set("Cache-Control", "no-cache")
		pagedata := BodyData{Value: "<pre><code>" + html.EscapeString(string(res)) + "</code></pre>"}
		tmpl := template.Must(template.ParseFiles(wd + "/backend/templates/layout.html"))
		err = tmpl.Execute(w, pagedata)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func notFoundHandler(w http.ResponseWriter) {
	// read these later once at init
	files, err := os.ReadDir(goPaste404Dir)
	if err != nil {
		fmt.Println(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Cache-Control", "no-cache")
	pagedata := BodyData{Value: "<div id='float404'><img src='/public/404/" + files[rand.Intn(len(files))].Name() + "'/></div>"} // #nosec G404
	tmpl := template.Must(template.ParseFiles(wd + "/backend/templates/layout.html"))
	err = tmpl.Execute(w, pagedata)
	if err != nil {
		fmt.Println(err)
	}
}

func getIndexHandler(w http.ResponseWriter) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Cache-Control", "no-cache")
	pagedata := BodyData{Value: "<textarea class='prettyprint' id='paste' placeholder='[ paste text  -  ctrl+s to save ]' spellcheck='false'></textarea>"}
	tmpl := template.Must(template.ParseFiles(wd + "/backend/templates/layout.html"))
	err = tmpl.Execute(w, pagedata)
	if err != nil {
		fmt.Println(err)
	}
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

func getDBType(s string) int {
	switch s {
	case "bolt":
		return 0
	case "memory":
		return 1
	default:
		return 0
	}
}

func initPyroscope(addr string) {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "github.com.ruupert.paste",
		ServerAddress:   addr,
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if goPastePyroscope != "" {
		initPyroscope(goPastePyroscope)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(requestHandler))
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./backend/public"))))
	d, err := pastedb.NewDatabaseType(pastedb.DatabaseType(getDBType(goPasteDb)))
	if err != nil {
		log.Fatal(err)
	}
	db = d
	if certsExist() {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS13,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
			},
		}
		srv := &http.Server{
			Addr:              fmt.Sprintf("%s:%d", goPasteAddr, goPastePort),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      5 * time.Second,
			MaxHeaderBytes:    8192,
			TLSConfig:         cfg,
			TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		fmt.Printf("Listening https on %d with %s store backend\n", goPastePort, db.GetName())
		log.Fatal(srv.ListenAndServeTLS(goPasteTlsCrt, goPasteTlsKey))
	} else {
		srv := &http.Server{
			Addr:              fmt.Sprintf("%s:%d", goPasteAddr, goPastePort),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      5 * time.Second,
			MaxHeaderBytes:    8192,
		}
		fmt.Printf("Listening http on %d with %s store backend\n", goPastePort, db.GetName())
		log.Fatal(srv.ListenAndServe())
	}
}