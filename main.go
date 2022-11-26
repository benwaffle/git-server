package main

import (
	"io"
	"log"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
)

func main() {
	http.HandleFunc("/info/refs", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s", r.RemoteAddr)
		log.Printf("\t%s", r.URL)
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

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("body: [%s]\n", buf)

		w.Header().Add("Content-Type", "application/x-git-upload-pack-advertisement")
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(200)

		pkt := pktline.NewEncoder(w)
		pkt.Encodef("# service=git-upload-pack\n")
		pkt.Flush()
		pkt.Encodef("7777777777777777777777777777777777777777 refs/heads/wheeee\x00%s\n", capability.Sideband64k)
		pkt.Flush()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("404 - %s (from %s)", r.URL, r.RemoteAddr)
		w.WriteHeader(404)
	})

	http.ListenAndServe(":5050", nil)
}
