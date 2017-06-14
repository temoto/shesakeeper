package main

import (
	"flag"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type Nothing struct{}

func getFileOwnership(path string) (uint32, uint32) {
	if info, err := os.Stat(path); err != nil {
		log.Fatal(err)
		panic("should not execute this line")
	} else {
		st := info.Sys().(*syscall.Stat_t)
		return st.Uid, st.Gid
	}
}

func main() {
	flag.Parse()
	root, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	_, rootGroup := getFileOwnership(root)

	events := make(chan notify.EventInfo, 4<<10)
	// TODO: maybe no need in Write events?
	if err := notify.Watch(filepath.Join(root, "..."), events, notify.Create, notify.Write); err != nil {
		log.Fatal(err)
	}

	for ev := range events {
		path, err := filepath.Abs(ev.Path())
		if err != nil {
			log.Fatal(err)
		}
		if !filepath.HasPrefix(path, root) {
			log.Fatal("hijack attempted path:", ev.Path(), "abs:", path)
		}
		// debug
		// log.Println("watch event:", ev)
		if !IsInterestingEvent(ev.Event()) {
			continue
		}
		if err = os.Chown(path, -1, int(rootGroup)); err != nil {
			log.Println("chown fail path:", path, "error:", err)
		}
	}
}
