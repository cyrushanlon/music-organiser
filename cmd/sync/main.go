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

//TODO: make it more directory aware

type fileInfo struct {
	Size int64
	Hash string
	Path string
}

var originPath = "/Users/cyrushanlon/Documents/Music"
var remotePath = "/Volumes/THE CYPOD/Music"

var wg = sync.WaitGroup{}

func process(m map[string]fileInfo, rootPathLen int, path string, linkPath string) {
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
		m[strings.ToLower(filePath[rootPathLen:])] = fileInfo{
			Path: filePath[rootPathLen:],
			Size: v.Size(),
			// Hash:  hash,
		}
	}
}

func getFileList(m map[string]fileInfo, path string) {
	process(m, len(path), path, "")
	wg.Done()
}

func main() {

	wg.Add(2)

	//build file list in origin
	originMap := make(map[string]fileInfo)
	go func() {
		s := time.Now()
		fmt.Println("Building local file list...")

		getFileList(originMap, originPath)

		fmt.Printf("Completed local in %s.\n", time.Since(s))
	}()

	// build file list in remote
	remoteMap := make(map[string]fileInfo)
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
			//fix casing if incorrect
			file.FixCasing(o.Path, r.Path, originPath, remotePath)

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
			//if the paths cases dont match rename and continue

			// this file has been deleted at the origin
			fmt.Println("Delete ", remotePath+k)

			err := os.Remove(remotePath + k)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, v := range originMap {
		fmt.Println("Copy   ", originPath+v.Path)

		remoteSplit := strings.Split(remotePath+v.Path, "/")

		err := os.MkdirAll(strings.Join(remoteSplit[:len(remoteSplit)-1], "/"), 0777)
		if err != nil {
			log.Fatal(err)
		}

		//get the parent directory of both and make sure the case is ok

		err = file.Copy(originPath+v.Path, remotePath+v.Path)
		if err != nil {
			log.Fatal(err)
		}
	}

	//cleanup empty folders

	fmt.Println("Done!")
}
