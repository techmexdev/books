package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kjk/u"
)

func copyFile(dst, src string) error {
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fout.Close()
	_, err = io.Copy(fout, fin)
	return err
}

func copyToWww(path string) {
	name := filepath.Base(path)
	dst := filepath.Join("www", name)
	err := copyFile(dst, path)
	u.PanicIfErr(err)
}

func rebuildAll() {
	// TOOD: implement me
}

func handleFileChange(path string) {
	fmt.Printf("handleFileChange: %s\n", path)
	if strings.HasSuffix(path, "main.css") {
		copyToWww(filepath.Join("tmpl", "main.css"))
		return
	}
	if strings.HasSuffix(path, ".tmpl.html") {
		rebuildAll()
		return
	}
	// TODO: if only .md or renamed directory, only rebuild that one book
}

func rebuildOnChanges() {
	dirs, err := getDirsRecur("tmpl")
	u.PanicIfErr(err)
	dirs2, err := getDirsRecur("books")
	u.PanicIfErr(err)
	dirs = append(dirs, dirs2...)

	watcher, err := fsnotify.NewWatcher()
	u.PanicIfErr(err)
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered in rebuildOnChanges(). Error: '%s'\n", r)
				// TODO: why this doesn't seem to trigger done
				done <- true
			}
		}()

		for {
			select {
			case event := <-watcher.Events:
				// filter out events that are just chmods
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				fmt.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("modified file:", event.Name)
				}
				handleFileChange(event.Name)
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()
	for _, dir := range dirs {
		fmt.Printf("Watching dir: '%s'\n", dir)
		watcher.Add(dir)
	}
	// waiting forever
	// TODO: pick up ctrl-c and cleanup and quit
	<-done
	fmt.Printf("exiting rebuildOnChanges()")
}
