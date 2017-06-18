package main

import (
	"flag"
	"github.com/coreos/go-systemd/daemon"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func getFileOwnership(path string) (uint32, uint32) {
	if info, err := os.Stat(path); err != nil {
		log.Fatal(err)
		panic("should not execute this line")
	} else {
		st := info.Sys().(*syscall.Stat_t)
		return st.Uid, st.Gid
	}
}

func sdnotify(s string) {
	if _, err := daemon.SdNotify(false, s); err != nil {
		log.Fatal(err)
	}
}

func main() {
	sdnotify("READY=0\nSTATUS=init\n")
	if wdTime, err := daemon.SdWatchdogEnabled(true); err != nil {
		log.Fatal(err)
	} else if wdTime != 0 {
		go func() {
			for _ = range time.Tick(wdTime) {
				sdnotify("WATCHDOG=1\n")
			}
		}()
	}

	flag.Parse()
	var root string
	var err error
	if root, err = filepath.Abs(flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
	if root, err = filepath.EvalSymlinks(root); err != nil {
		log.Fatal(err)
	}
	_, keepGroupId := getFileOwnership(root)

	events := make(chan notify.EventInfo, 4<<10)
	// TODO: maybe no need in Write events?
	if err := notify.Watch(filepath.Join(root, "..."), events, notify.Create, notify.Write); err != nil {
		log.Fatal(err)
	}

	sdnotify("READY=1\nSTATUS=work\n")
	keepGroupName := ""
	if keepGroup, err := user.LookupGroupId(strconv.Itoa(int(keepGroupId))); err != nil {
		log.Fatal(err)
	} else {
		keepGroupName = keepGroup.Name
	}
	log.Printf("status=work root=%s group=%s", root, keepGroupName)

	for ev := range events {
		path, err := filepath.Abs(ev.Path())
		if err != nil {
			log.Fatal(err)
		}
		if !filepath.HasPrefix(path, root) {
			log.Fatalf("hijack attempted path: %s abs: %s is not under root: %s", ev.Path(), path, root)
		}
		// debug
		// log.Println("watch event:", ev)
		if !IsInterestingEvent(ev.Event()) {
			continue
		}
		if err = os.Chown(path, -1, int(keepGroupId)); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("chown fail path: %s error: %s", path, err.Error())
			}
		}
	}
}
