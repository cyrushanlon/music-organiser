package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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
