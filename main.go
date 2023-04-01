package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type RsyncConfig struct {
	RemoteHost   string
	RemotePath   string
	RemoteUser   string
	RemotePass   string
	MirrorFolder string
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if !isGitRepo(currentDir) {
		log.Fatal("Current directory is not a git repository")
	}

	rsyncConfig, err := loadRsyncConfig(".rsync")
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if !strings.HasSuffix(event.Name, ".git/") {
						fmt.Printf("File changed: %s\n", event.Name)
						sync_remote(event.Name, rsyncConfig)
					}
				}
			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return !os.IsNotExist(err)
}

func loadRsyncConfig(fileName string) (*RsyncConfig, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &RsyncConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			switch key {
			case "RemoteHost":
				config.RemoteHost = value
			case "RemotePath":
				config.RemotePath = value
			case "RemoteUser":
				config.RemoteUser = value
			case "RemotePass":
				config.RemotePass = value
			case "MirrorFolder":
				config.MirrorFolder = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func sync_remote(filePath string, config *RsyncConfig) {
	// Implement your SFTP or other remote file synchronization logic here.
	fmt.Printf("Syncing remote file: %s\n", filePath)
	fmt.Println("Remote host:", config.RemoteHost)
	fmt.Println("Remote path:", config.RemotePath)
	fmt.Println("Remote user:", config.RemoteUser)
	// Don't print the password for security reasons
	fmt.Println("Mirror folder:", config.MirrorFolder)
}

