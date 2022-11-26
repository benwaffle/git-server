package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/muesli/termenv"
)

type SidebandTTY struct {
	mux *sideband.Muxer
}

func (s SidebandTTY) Write(p []byte) (n int, err error) {
	return s.mux.WriteChannel(sideband.ProgressMessage, p)
}

func flushHttp(w http.ResponseWriter) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func main() {
	http.HandleFunc("/info/refs", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		for k, v := range r.Header {
			log.Printf("\t%s: %s\n", k, v)
		}
		if r.Header.Get("git-protocol") != "version=2" {
			log.Printf("bad protocol")
			w.WriteHeader(400)
			return
		}
		if r.URL.Query()["service"][0] != "git-upload-pack" {
			log.Printf("bad service")
			w.WriteHeader(400)
			return
		}

		w.Header().Add("Content-Type", "application/x-git-upload-pack-advertisement")
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(200)

		pkt := pktline.NewEncoder(w)
		pkt.Encodef("version 2\n")
		pkt.Encodef("ls-refs\n")
		pkt.Encodef("fetch\n")
		pkt.Flush()
	})

	http.HandleFunc("/git-upload-pack", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		w.Header().Add("Content-Type", "application/x-git-upload-pack-advertisement")
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(200)

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		body := string(buf)
		log.Printf("body: [%s]\n", body)

		if strings.Contains(body, "command=ls-refs") {
			pkt := pktline.NewEncoder(w)
			pkt.Encodef("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef refs/heads/wheeee\n")
			pkt.Flush()
		} else if strings.Contains(body, "command=fetch") {
			pkt := pktline.NewEncoder(w)
			pkt.Encodef("packfile\n")

			mux := sideband.NewMuxer(sideband.Sideband, w)

			faketty := SidebandTTY{mux: mux}
			output := termenv.NewOutput(faketty, termenv.WithProfile(termenv.TrueColor))

			output.ClearScreen()
			output.MoveCursor(10, 10)
			output.WriteString(
				output.String("hello git").Bold().Foreground(output.Color("#abcdef")).String(),
			)
			output.CursorDown(10)
			output.WriteString("\n")
			flushHttp(w)
			time.Sleep(1 * time.Second)

			output.ClearScreen()
			output.WriteString(
				output.String("bye git").Bold().Foreground(output.Color("#abcdef")).String(),
			)
			output.CursorDown(10)
			output.WriteString("\n")
			flushHttp(w)
			time.Sleep(1 * time.Second)

			// for i := 0; i <= 200; i += 1 {
			// 	mux.WriteChannel(sideband.ProgressMessage, []byte(fmt.Sprintf("%d\n", i)))

			// 	if f, ok := w.(http.Flusher); ok {
			// 		log.Printf("flushing - %d", i)
			// 		f.Flush()
			// 	}
			// 	time.Sleep(10 * time.Millisecond)
			// }

			pkt.Flush()

			pkt.Encodef("done\n")
			pkt.Flush()
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("404 - %s (from %s)", r.URL, r.RemoteAddr)
		w.WriteHeader(404)
	})

	log.Printf("listening on :5050")
	http.ListenAndServe(":5050", nil)
}
