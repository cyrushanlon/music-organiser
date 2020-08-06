package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func GetAll(p string) []os.FileInfo {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func Hash(path string) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return "", nil
	}
	defer fh.Close()

	h := sha256.New()
	_, err = io.Copy(h, fh)
	if err != nil {
		return "", nil
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func Copy(src string, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

//IsEmpty returns true if a given directory is empty, false otherwise
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// FixCasing ensures the casing of all files/folders in path remote are equal to that of path origin
// applies the prefixes originPath to origin and remotePath to remote
func FixCasing(origin, remote, originPath, remotePath string) {
	splitOrigin := strings.Split(origin, "/")
	splitRemote := strings.Split(remote, "/")
	for i := len(splitOrigin) - 1; i > 0; i-- {
		if splitOrigin[i] != splitRemote[i] {
			fmt.Printf("Rename: \"%s\" to \"%s\"\n", splitRemote[i], splitOrigin[i])
			//build up path to get to the thing that needs renaming
			p := splitRemote[i]
			r := splitOrigin[i]
			for j := i - 1; j > 0; j-- {
				p = splitRemote[j] + "/" + p
				r = splitOrigin[j] + "/" + r
			}
			// fmt.Println(p)
			err := os.Rename(remotePath+"/"+p, remotePath+"/"+r)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
