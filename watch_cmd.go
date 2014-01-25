package main

import (
	"bufio"
	"code.google.com/p/go.exp/inotify"
	"log"
	"os"
	"time"
)

var watchEntryCmd = &Command{
	Name:    "watch",
	Usage:   "an internal command",
	Summary: "used to watch a journal entry until it has been completed and saved",
	Help:    "do not use",
	Run: func(c *Command, args ...string) {
		_, err := watchEntry(nil, c, args...)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func watchEntry(statusCh chan<- string, c *Command, args ...string) (t time.Time, err error) {
	t = time.Now()

	watcher, err := inotify.NewWatcher()
	if err != nil {
		return t, err
	}

	filename := args[0]

	err = watcher.Watch(filename)
	if err != nil {
		return t, err
	}

	defer watcher.Close()

	for {
		select {
		case statusCh <- "watching":
		case err := <-watcher.Error:
			return t, err
		case e := <-watcher.Event:
			if e.Mask&inotify.IN_CLOSE != 0 {
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0600)
				if err != nil {
					return t, err
				}

				defer f.Close()
				fbuf := bufio.NewWriter(f)

				t = time.Now()
				fbuf.WriteString(t.Format(time.UnixDate) + "\n")

				err = fbuf.Flush()
				if err != nil {
					return t, err
				}

				return t, nil
			}
		}
	}
	return t, nil
}
