package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	bind          string
	port          int
	baseDirectory string
)

func init() {
	flag.StringVar(&bind, "bind", "0.0.0.0", "bind address, default: all interfaces")
	flag.IntVar(&port, "port", 8000, "bind port, default: 8000")
	flag.StringVar(&baseDirectory, "directory", ".", "base directory, default: current directory")
	flag.Parse()
	baseDirectory, _ = filepath.Abs(baseDirectory)
	log.Printf("bind: %s, port: %d, directory: %s", bind, port, baseDirectory)
}

func main() {
	http.HandleFunc("/", serveGet)

	addr := net.JoinHostPort(bind, strconv.Itoa(port))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("getting %s", r.URL.String())
	fullPath := localPath(r.URL.Path)

	stat, err := os.Stat(fullPath)
	if err != nil {
		log.Printf("Error: %v", err)
		switch err {
		case os.ErrNotExist:
			w.WriteHeader(http.StatusNotFound)
		case os.ErrPermission:
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if stat.IsDir() {
		listDir(w, r, fullPath)
	} else {
		sendFile(w, fullPath)
	}
}

func localPath(path string) string {
	parts := strings.Split(path, "/")
	return filepath.Join(append([]string{baseDirectory}, parts...)...)
}

func listDir(w http.ResponseWriter, r *http.Request, fullPath string) {
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Error: %v", err)
		switch err {
		case os.ErrNotExist:
			w.WriteHeader(http.StatusNotFound)
		case os.ErrPermission:
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	infos, err := file.Readdir(0)
	if err != nil {
		log.Printf("Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buffer := bytes.Buffer{}
	buffer.WriteString(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">`)
	buffer.WriteString("\n<html>\n<head>")
	buffer.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n</head>\n")
	buffer.WriteString("<body>\n<hr>\n<ul>\n")

	for _, info := range infos {
		name := info.Name()
		if info.IsDir() {
			name += "/"
		}
		link := path.Join(r.URL.Path, url.PathEscape(name))
		buffer.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", link, html.EscapeString(name)))
	}

	buffer.WriteString("</ul>\n<hr>\n</body>\n</html>\n")
	w.Write(buffer.Bytes())
}

func sendFile(w http.ResponseWriter, fullPath string) {
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Error: %v", err)
		switch err {
		case os.ErrNotExist:
			w.WriteHeader(http.StatusNotFound)
		case os.ErrPermission:
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	buffer := make([]byte, 4096)
	io.CopyBuffer(w, file, buffer)
}
