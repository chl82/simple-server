package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
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
	extToType     map[string]string
)

func init() {
	flag.StringVar(&bind, "bind", "0.0.0.0", "bind address, default: all interfaces")
	flag.IntVar(&port, "port", 8000, "bind port, default: 8000")
	flag.StringVar(&baseDirectory, "directory", ".", "base directory, default: current directory")
	flag.Parse()
	baseDirectory, _ = filepath.Abs(baseDirectory)
	log.Printf("bind: %s, port: %d, directory: %s", bind, port, baseDirectory)

	extToType = map[string]string{
		".3fr":                "image/3fr",
		".3g2":                "video/3gpp2",
		".3gp":                "video/3gpp",
		".3gp2":               "video/3gpp2",
		".3gpp":               "video/3gpp",
		".wmd":                "application/x-ms-wmd",
		".a":                  "application/octet-stream",
		".aac":                "audio/vnd.dlna.adts",
		".ac3":                "audio/vnd.dolby.dd-raw",
		".accountpicture-ms":  "application/windows-accountpicture",
		".adt":                "audio/vnd.dlna.adts",
		".adts":               "audio/vnd.dlna.adts",
		".ai":                 "application/postscript",
		".aif":                "audio/aiff",
		".aifc":               "audio/aiff",
		".aiff":               "audio/aiff",
		".appcontent-ms":      "application/windows-appcontent+xml",
		".application":        "application/x-ms-application",
		".ari":                "image/ari",
		".arw":                "image/arw",
		".asf":                "video/x-ms-asf",
		".asx":                "video/x-ms-asf",
		".au":                 "audio/basic",
		".avi":                "video/avi",
		".bat":                "text/plain",
		".bay":                "image/bay",
		".bcpio":              "application/x-bcpio",
		".bin":                "application/octet-stream",
		".bmp":                "image/bmp",
		".c":                  "text/plain",
		".cap":                "image/cap",
		".cat":                "application/vnd.ms-pki.seccat",
		".cdf":                "application/x-netcdf",
		".cer":                "application/x-x509-ca-cert",
		".contact":            "text/x-ms-contact",
		".cpio":               "application/x-cpio",
		".cr2":                "image/cr2",
		".cr3":                "image/cr3",
		".crl":                "application/pkix-crl",
		".crt":                "application/x-x509-ca-cert",
		".crw":                "image/crw",
		".csh":                "application/x-csh",
		".css":                "text/css",
		".csv":                "text/csv",
		".dcr":                "image/dcr",
		".dcs":                "image/dcs",
		".dds":                "image/vnd.ms-dds",
		".der":                "application/x-x509-ca-cert",
		".dib":                "image/bmp",
		".dll":                "application/x-msdownload",
		".dng":                "image/dng",
		".doc":                "application/msword",
		".dot":                "application/msword",
		".drf":                "image/drf",
		".dtcp-ip":            "application/x-dtcp1",
		".dvi":                "application/x-dvi",
		".dvr-ms":             "video/x-ms-dvr",
		".ec3":                "audio/ec3",
		".eip":                "image/eip",
		".emf":                "image/x-emf",
		".eml":                "message/rfc822",
		".eps":                "application/postscript",
		".epub":               "application/epub+zip",
		".erf":                "image/erf",
		".etx":                "text/x-setext",
		".exe":                "application/x-msdownload",
		".fff":                "image/fff",
		".fif":                "application/fractals",
		".flac":               "audio/x-flac",
		".gif":                "image/gif",
		".group":              "text/x-ms-group",
		".gtar":               "application/x-gtar",
		".gz":                 "application/x-gzip",
		".h":                  "text/plain",
		".hdf":                "application/x-hdf",
		".hqx":                "application/mac-binhex40",
		".hta":                "application/hta",
		".htc":                "text/x-component",
		".htm":                "text/html",
		".html":               "text/html",
		".ico":                "image/x-icon",
		".ief":                "image/ief",
		".iiq":                "image/iiq",
		".jfif":               "image/jpeg",
		".jpe":                "image/jpeg",
		".jpeg":               "image/jpeg",
		".jpg":                "image/jpeg",
		".js":                 "application/javascript",
		".json":               "application/json",
		".jxr":                "image/vnd.ms-photo",
		".k25":                "image/k25",
		".kdc":                "image/kdc",
		".ksh":                "text/plain",
		".latex":              "application/x-latex",
		".library-ms":         "application/windows-library+xml",
		".lpcm":               "audio/l16",
		".m1v":                "video/mpeg",
		".m2t":                "video/vnd.dlna.mpeg-tts",
		".m2ts":               "video/vnd.dlna.mpeg-tts",
		".m2v":                "video/mpeg",
		".m3u":                "audio/x-mpegurl",
		".m3u8":               "application/vnd.apple.mpegurl",
		".m4a":                "audio/mp4",
		".m4v":                "video/mp4",
		".man":                "application/x-troff-man",
		".me":                 "application/x-troff-me",
		".mef":                "image/mef",
		".mht":                "message/rfc822",
		".mhtml":              "message/rfc822",
		".mid":                "audio/mid",
		".midi":               "audio/mid",
		".mif":                "application/x-mif",
		".mjs":                "application/javascript",
		".mka":                "audio/x-matroska",
		".mkv":                "video/x-matroska",
		".mod":                "video/mpeg",
		".mos":                "image/mos",
		".mov":                "video/quicktime",
		".movie":              "video/x-sgi-movie",
		".mp2":                "audio/mpeg",
		".mp2v":               "video/mpeg",
		".mp3":                "audio/mpeg",
		".mp4":                "video/mp4",
		".mp4v":               "video/mp4",
		".mpa":                "audio/mpeg",
		".mpe":                "video/mpeg",
		".mpeg":               "video/mpeg",
		".mpg":                "video/mpeg",
		".mpv2":               "video/mpeg",
		".mrw":                "image/mrw",
		".ms":                 "application/x-troff-ms",
		".msepub":             "application/epub+zip",
		".mts":                "video/vnd.dlna.mpeg-tts",
		".nc":                 "application/x-netcdf",
		".nef":                "image/nef",
		".nrw":                "image/nrw",
		".nws":                "message/rfc822",
		".o":                  "application/octet-stream",
		".obj":                "application/octet-stream",
		".oda":                "application/oda",
		".oga":                "audio/ogg",
		".ogg":                "audio/ogg",
		".ogm":                "video/ogg",
		".ogv":                "video/ogg",
		".ogx":                "video/ogg",
		".opus":               "audio/ogg",
		".orf":                "image/orf",
		".ori":                "image/cr3",
		".osdx":               "application/opensearchdescription+xml",
		".p10":                "application/pkcs10",
		".p12":                "application/x-pkcs12",
		".p7b":                "application/x-pkcs7-certificates",
		".p7c":                "application/pkcs7-mime",
		".p7m":                "application/pkcs7-mime",
		".p7r":                "application/x-pkcs7-certreqresp",
		".p7s":                "application/pkcs7-signature",
		".pbm":                "image/x-portable-bitmap",
		".pdf":                "application/pdf",
		".pef":                "image/pef",
		".pfx":                "application/x-pkcs12",
		".pgm":                "image/x-portable-graymap",
		".pko":                "application/vnd.ms-pki.pko",
		".pl":                 "text/plain",
		".png":                "image/png",
		".pnm":                "image/x-portable-anymap",
		".pot":                "application/vnd.ms-powerpoint",
		".ppa":                "application/vnd.ms-powerpoint",
		".ppm":                "image/x-portable-pixmap",
		".pps":                "application/vnd.ms-powerpoint",
		".ppt":                "application/vnd.ms-powerpoint",
		".prf":                "application/pics-rules",
		".ps":                 "application/postscript",
		".psc1":               "application/powershell",
		".ptx":                "image/ptx",
		".pwz":                "application/vnd.ms-powerpoint",
		".pxn":                "image/pxn",
		".py":                 "text/plain",
		".pyc":                "application/x-python-code",
		".pyo":                "application/x-python-code",
		".pyw":                "text/plain",
		".pyz":                "application/x-zip-compressed",
		".pyzw":               "application/x-zip-compressed",
		".qt":                 "video/quicktime",
		".ra":                 "audio/x-pn-realaudio",
		".raf":                "image/raf",
		".ram":                "application/x-pn-realaudio",
		".ras":                "image/x-cmu-raster",
		".rat":                "application/rat-file",
		".raw":                "image/raw",
		".rdf":                "application/xml",
		".rgb":                "image/x-rgb",
		".rmi":                "audio/mid",
		".roff":               "application/x-troff",
		".rtx":                "text/richtext",
		".rw2":                "image/rw2",
		".rwl":                "image/rwl",
		".sct":                "text/scriptlet",
		".searchconnector-ms": "application/windows-search-connector+xml",
		".sgm":                "text/x-sgml",
		".sgml":               "text/x-sgml",
		".sh":                 "application/x-sh",
		".shar":               "application/x-shar",
		".sit":                "application/x-stuffit",
		".snd":                "audio/basic",
		".so":                 "application/octet-stream",
		".sol":                "text/plain",
		".sor":                "text/plain",
		".spc":                "application/x-pkcs7-certificates",
		".spl":                "application/futuresplash",
		".sr2":                "image/sr2",
		".src":                "application/x-wais-source",
		".srf":                "image/srf",
		".srw":                "image/srw",
		".sst":                "application/vnd.ms-pki.certstore",
		".sv4cpio":            "application/x-sv4cpio",
		".sv4crc":             "application/x-sv4crc",
		".svg":                "image/svg+xml",
		".swf":                "application/x-shockwave-flash",
		".t":                  "application/x-troff",
		".tar":                "application/x-tar",
		".tcl":                "application/x-tcl",
		".tex":                "application/x-tex",
		".texi":               "application/x-texinfo",
		".texinfo":            "application/x-texinfo",
		".tgz":                "application/x-compressed",
		".tif":                "image/tiff",
		".tiff":               "image/tiff",
		".tod":                "video/mpeg",
		".tr":                 "application/x-troff",
		".ts":                 "video/vnd.dlna.mpeg-tts",
		".tsv":                "text/tab-separated-values",
		".tts":                "video/vnd.dlna.mpeg-tts",
		".txt":                "text/plain",
		".ustar":              "application/x-ustar",
		".uvu":                "video/vnd.dece.mp4",
		".vcf":                "text/x-vcard",
		".wasm":               "application/wasm",
		".wav":                "audio/wav",
		".wax":                "audio/x-ms-wax",
		".wdp":                "image/vnd.ms-photo",
		".weba":               "audio/webm",
		".webm":               "video/webm",
		".webmanifest":        "application/manifest+json",
		".website":            "application/x-mswebsite",
		".wiz":                "application/msword",
		".wm":                 "video/x-ms-wm",
		".wma":                "audio/x-ms-wma",
		".wmf":                "image/x-wmf",
		".wmv":                "video/x-ms-wmv",
		".wmx":                "video/x-ms-wmx",
		".wmz":                "application/x-ms-wmz",
		".wpl":                "application/vnd.ms-wpl",
		".wsc":                "text/scriptlet",
		".wsdl":               "application/xml",
		".wvx":                "video/x-ms-wvx",
		".x3f":                "image/x3f",
		".xaml":               "application/xaml+xml",
		".xbap":               "application/x-ms-xbap",
		".xbm":                "image/x-xbitmap",
		".xht":                "application/xhtml+xml",
		".xhtml":              "application/xhtml+xml",
		".xlb":                "application/vnd.ms-excel",
		".xls":                "application/vnd.ms-excel",
		".xml":                "text/xml",
		".xpdl":               "application/xml",
		".xpm":                "image/x-xpixmap",
		".xrm-ms":             "text/xml",
		".xsl":                "text/xml",
		".xwd":                "image/x-xwindowdump",
		".z":                  "application/x-compress",
		".zip":                "application/x-zip-compressed",
	}
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

		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else if os.IsPermission(err) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
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
	infos, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Printf("Error: %v", err)

		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else if os.IsPermission(err) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))

	w.Write(buffer.Bytes())
}

func sendFile(w http.ResponseWriter, fullPath string) {
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Error: %v", err)

		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else if os.IsPermission(err) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	w.Header().Set("Content-Type", guessContentType(fullPath))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	buffer := make([]byte, 4096)
	io.CopyBuffer(w, file, buffer)
}

func guessContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	contentType, ok := extToType[ext]
	if ok {
		return contentType
	}
	return "application/octet-stream"
}
