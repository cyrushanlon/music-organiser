package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cyrushanlon/music-organiser/file"
	"golang.org/x/text/unicode/norm"
)

//TODO: handle errors
//TODO: CLI
//TODO: add toggle for special characters
//TODO: when directory case changes the album is deleted and copied perpetually

type FileInfo struct {
	IsDir bool

	Size int64
	Hash string
}

var originPath = "/Users/cyrushanlon/Documents/Music"
var remotePath = "/Volumes/THE CYPOD/Music"

var wg = sync.WaitGroup{}

func process(m map[string]FileInfo, rootPathLen int, path string, linkPath string) {
	if linkPath == "" {
		linkPath = path
	}

	f := file.GetAll(path)
	for _, v := range f {
		if v.Name() == ".DS_Store" {
			continue
		}

		//normalise the names because ipod is fat32
		filePath := path + "/" + norm.NFC.String(v.Name())

		if v.IsDir() {
			process(m, rootPathLen, filePath, filePath)
			continue
		}
		//could be synbolic link?
		if v.Mode().String()[0] == 'L' {
			p, _ := filepath.EvalSymlinks(filePath)
			process(m, rootPathLen, filePath, p)
			continue
		}
		//is a file
		//get hash
		// hash, _ := file.Hash(filePath)
		m[filePath[rootPathLen:]] = FileInfo{
			IsDir: false,
			Size:  v.Size(),
			// Hash:  hash,
		}
	}
}

func getFileList(m map[string]FileInfo, path string) {
	process(m, len(path), path, "")
	wg.Done()
}

func main() {

	wg.Add(2)

	//build file list in origin
	originMap := make(map[string]FileInfo)
	go func() {
		s := time.Now()
		fmt.Println("Building local file list...")

		getFileList(originMap, originPath)

		fmt.Printf("Completed local in %s.\n", time.Since(s))
	}()

	// build file list in remote
	remoteMap := make(map[string]FileInfo)
	go func() {
		s := time.Now()
		fmt.Println("Building remote file list...")

		getFileList(remoteMap, remotePath)

		fmt.Printf("Completed remote in %s.\n", time.Since(s))
	}()

	wg.Wait()

	fmt.Println("Finding Diffs")

	for k, r := range remoteMap {
		if o, ok := originMap[k]; ok {
			//the file name exists at both, does the hash match?
			if o.Size != r.Size {
				// the file has changed at the origin
				fmt.Println("Replace", remotePath+k)

				err := os.Remove(remotePath + k)
				if err != nil {
					log.Fatal(err)
				}

				err = file.Copy(originPath+k, remotePath+k)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				//we can remove this file from the origin map now
				delete(originMap, k)
			}
		} else {
			// this file has been deleted at the origin
			fmt.Println("Delete ", remotePath+k)

			err := os.Remove(remotePath + k)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for k := range originMap {
		fmt.Println("Copy  ", originPath+k)

		split := strings.Split(remotePath+k, "/")
		err := os.MkdirAll(strings.Join(split[:len(split)-1], "/"), 0777)
		if err != nil {
			log.Fatal(err)
		}

		err = file.Copy(originPath+k, remotePath+k)
		if err != nil {
			log.Fatal(err)
		}
	}

	//cleanup empty folders

	fmt.Println("Done!")
}
