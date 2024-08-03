//go:build !wasm
// +build !wasm

package main

//go:generate ./generators/front.sh
import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"embed"

	pastedb "github.com/ruupert/paste/db"
)

type BodyData struct {
	Value string
}

var (
	db                    pastedb.DatabaseInterface
	goPastePort           int
	goPasteAddr           string
	goPasteTlsCrt         string
	goPasteTlsKey         string
	goPasteDb             string
	goPastePyroscope      string
	goPastePyroscopePort  string
	goPastePyroscopeProto string
)

//go:embed assets/out.wasm
//go:embed assets/script/wasm_exec.js
//go:embed assets/css/paste.css
//go:embed assets/404/404_1.png
//go:embed assets/templates/layout.html
var embedded embed.FS

func init() {
	flag.IntVar(&goPastePort, "port", 8432, "Listen port")
	flag.StringVar(&goPasteAddr, "addr", "0.0.0.0", "Bind address")
	flag.StringVar(&goPasteTlsCrt, "cert", "tls.pem", "Cert")
	flag.StringVar(&goPasteTlsKey, "key", "tls.key", "Cert Key")
	flag.StringVar(&goPasteDb, "db", "bolt", "backend options: [bolt, memory]")
	flag.StringVar(&goPastePyroscope, "pyroscope", "", "Pyroscope ServerAddress")
	flag.StringVar(&goPastePyroscopePort, "pyroscope port", "4040", "Pyroscope port")
	flag.StringVar(&goPastePyroscopeProto, "pyroscope proto", "http", "Pyroscope proto")
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
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/pyroscope":
		if goPastePyroscope != "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(fmt.Sprintf("%s://%s:%s", goPastePyroscopeProto, goPastePyroscope, goPastePyroscopePort)))
			if err != nil {
				fmt.Println(err)
			}
			return
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			_, err := w.Write([]byte(""))
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	case "/wasssm":
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFileFS(w, r, embedded, "assets/out.wasm")
	case "/css":
		w.Header().Set("Content-Type", "text/css")
		http.ServeFileFS(w, r, embedded, "assets/css/paste.css")
		return
	case "/wasmexec":
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFileFS(w, r, embedded, "assets/script/wasm_exec.js")
		return
	case "/404":
		w.Header().Set("Content-Type", "image/png")
		http.ServeFileFS(w, r, embedded, "assets/404/404_1.png")
		return
	default:
		uvals := r.URL.Query()
		if uvals.Has("q") {
			q := uvals.Get("q")
			res, err := db.Get([]byte(q))
			if err != nil {
				fmt.Println("err get")
				notFoundHandler(w)
				return
			}
			fmt.Println(res)
			_, err = w.Write(res)
			if err != nil {
				fmt.Println(err)
			}
			return
		} else {
			getIndexHandler(w)
			return
		}
	}
}

func notFoundHandler(w http.ResponseWriter) {
	http.Error(w, "", http.StatusNotFound)
}

func getIndexHandler(w http.ResponseWriter) {

	w.Header().Set("Cache-Control", "no-cache")
	pagedata := BodyData{Value: ""}
	tmpl := template.Must(template.ParseFS(embedded, "assets/templates/layout.html"))
	err := tmpl.Execute(w, pagedata)
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

/*
	func initPyroscope(addr string, proto string, port string) {
		runtime.SetMutexProfileFraction(5)
		runtime.SetBlockProfileRate(5)
		_, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: "github.com.ruupert.paste.backend",
			ServerAddress:   fmt.Sprintf("%s://%s:%s", proto, addr, port),
			Logger:          nil, // pyroscope.StandardLogger,
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
*/
func main() {
	/*if goPastePyroscope != "" {
		initPyroscope(goPastePyroscopeProto, goPastePyroscope, goPastePyroscopePort)
	}*/

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(requestHandler))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./public"))))
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
