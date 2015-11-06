package main

import (
	"log"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/exp/inotify"
)

func main() {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	flags := inotify.IN_CLOSE_WRITE | inotify.IN_MODIFY | inotify.IN_DELETE
	filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		// Only interested in directories
		if info != nil && !info.IsDir() {
			return nil
		}
		// Consume errors
		if err != nil {
			log.Println(err)
			return nil
		}
		// In case of error adding the watcher, stop everything
		err = watcher.AddWatch(path, flags)
		if err != nil {
			if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOSPC {
				log.Println("Too many watchers were set")
			} else {
				log.Println(err)
			}
			return err
		}
		return nil
	})
	log.Println("ready to listen to events...")
	for {
		select {
		case ev := <-watcher.Event:
			log.Println("event:", ev)
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}
