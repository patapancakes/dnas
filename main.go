package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/patapancakes/sslspoof"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("DNAS server emulator by Pancakes (pancakes@mooglepowered.com)")
	fmt.Println()

	us := flag.String("us", ":443", "address to bind us dnas server to")
	jp := flag.String("jp", "", "address to bind jp dnas server to")
	flag.Parse()

	mux := http.NewServeMux()

	// initialization
	mux.HandleFunc("POST /{region}/i-connect", handle) // unactivated hdds only?
	mux.HandleFunc("POST /{region}/d-connect", handle)

	// everything else
	mux.HandleFunc("POST /{region}/others", handle)

	var eg errgroup.Group
	for region, addr := range map[string]string{"us": *us, "jp": *jp} {
		if addr == "" {
			log.Println("skipping", region, "dnas server (address not set)")
			continue
		}

		log.Println("starting", region, "dnas server on", addr)

		l, err := sslspoof.NewListener(addr, "gate1."+region+".dnas.playstation.org", false)
		if err != nil {
			log.Fatalln("failed to start", region, "dnas server:", err)
		}

		defer l.Close()

		eg.Go(func() error { return http.Serve(l, mux) })
	}
	err := eg.Wait()
	if err != nil {
		log.Fatalln("dnas http server returned error:", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	req, err := readMessage(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := req
	resp.Data = []byte{}

	switch req.Opcode & 0xFF {
	case 0: // initialize
		resp.Opcode = req.Opcode&0xFF00 | 0x05
		if req.Flags&0x80 != 0 { // activated
			resp.Data = append(resp.Data, 0x00) // result
		}
		resp.Data = append(resp.Data, make([]byte, 25)...) // security
	case 1: // set hdd activation secret
		resp.Opcode = req.Opcode&0xFF00 | 0x06
		resp.Data = make([]byte, 6) // secret
	case 2: // install
		resp.Opcode = req.Opcode&0xFF00 | 0x0C
		resp.Data = make([]byte, 1) // result
	default:
		http.Error(w, "unhandled opcode", http.StatusBadRequest)
		return
	}

	resp.writeTo(w)
}
