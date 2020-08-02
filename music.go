package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

//TODO: handle multiple CD albums and duplicates
//TODO: do we care about "The Beatles" being "Beatles, The"?
//TODO: what year to use for compilations?
//TODO: delete source folders if everything worked
//TODO: CLI
//TODO: various folder confirmations

var (
	path    = "/Users/cyrushanlon/Documents/Music Copy"
	outPath = "/Users/cyrushanlon/Documents/Out"

	artists map[string]*artist
)

var (
	hideErrors = false
	dummy      = true
)

type artist struct {
	albums map[string]*album

	//holds the specific spelling of the artists name, most common is the one used
	appearedAs map[string]int
}

func (a *artist) Name() string {
	return getMostCommon(a.appearedAs)
}

type album struct {
	path string
	year string

	//holds the specific spelling of the albums name, most common is the one used
	appearedAs map[string]int
}

func (a *album) Name() string {
	return getMostCommon(a.appearedAs)
}

func getMostCommon(m map[string]int) string {
	max := 0
	name := ""
	for k, v := range m {
		if v > max {
			max = v
			name = k
		}
	}
	return name
}

func getFiles(p string) []os.FileInfo {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func processDirectory(p string) error {

	albumArtist := ""
	albumName := ""
	albumYear := 0

	var art *artist
	var ok bool

	artistAppearedAs := make(map[string]int)
	alb := &album{
		path:       p,
		appearedAs: make(map[string]int),
	}

	various := false

	for _, file := range getFiles(p) {
		if file.IsDir() {
			err := processDirectory(p + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		f, err := os.Open(p + "/" + file.Name())
		if err != nil {
			return err
		}

		m, err := tag.ReadFrom(f)
		if err != nil {
			f.Close()
			//if this file isnt a music file, we should keep crawling the folder
			continue
		}

		//read the tags from the track
		trackArtist := strings.Trim(strings.ReplaceAll(m.AlbumArtist(), "/", " "), " ")
		if trackArtist == "" {
			trackArtist = strings.Trim(strings.ReplaceAll(m.Artist(), "/", " "), " ")
			if trackArtist == "" {
				if !hideErrors {
					fmt.Println("missing name:", p+"/"+file.Name())
				}
				f.Close()
				return nil
			}
		}

		trackAlbum := strings.Trim(strings.ReplaceAll(m.Album(), "/", " "), " ")
		if trackAlbum == "" {
			if !hideErrors {
				fmt.Println("missing album:", p+"/"+file.Name())
			}
			f.Close()
			return nil
		}

		trackYear := m.Year()
		if trackYear == 0 {
			if !hideErrors {
				fmt.Println("missing year:", p+"/"+file.Name())
			}
			f.Close()
			return nil
		}

		if strings.TrimPrefix(m.Title(), " ") == "" {
			if !hideErrors {
				fmt.Println("missing track title:", p+"/"+file.Name())
			}
			f.Close()
			return nil
		}

		lowerTrackArtist := strings.ToLower(trackArtist)
		lowerTrackAlbum := strings.ToLower(trackAlbum)

		if !various {
			if albumArtist == "" {
				albumArtist = lowerTrackArtist
			} else if albumArtist != lowerTrackArtist {
				various = true
			}
			artistAppearedAs[trackArtist]++
		}

		mismatched := false
		if albumName == "" {
			albumName = lowerTrackAlbum
		} else if albumName != lowerTrackAlbum {
			//they should be the same in an album
			mismatched = true
		}
		alb.appearedAs[trackAlbum]++

		if albumYear == 0 {
			albumYear = trackYear
			alb.year = strconv.Itoa(trackYear)
		} else if albumYear != trackYear {
			//they should be the same in an album
			mismatched = true
		}

		if mismatched && !strings.Contains(strings.ToLower(m.Comment()), "compilation") {
			if !hideErrors {
				fmt.Println("mismatched data:", p+"/"+file.Name())
			}
			f.Close()
			return nil
		}

		f.Close()
	}

	if various {
		albumArtist = "various"
		artistAppearedAs = map[string]int{
			"Various": 1,
		}
	}

	if albumArtist != "" && albumName != "" && albumYear != 0 {
		if art, ok = artists[albumArtist]; !ok {
			art = &artist{
				albums:     make(map[string]*album),
				appearedAs: make(map[string]int),
			}
			artists[albumArtist] = art
		}

		if _, ok := art.albums[albumName]; ok {
			if !hideErrors {
				fmt.Println("duplicate album:", p)
			}
			return nil
		}

		art.albums[albumName] = alb

		for k, v := range artistAppearedAs {
			art.appearedAs[k] += v
		}
	}

	return nil
}

var fileCount = 0

func main() {

	artists = make(map[string]*artist)

	processDirectory(path)

	fmt.Println("processed")

	albumCount := 0
	fileCount := 0
	// now that we have the album/artist structure and info, we build the corresponding output
	os.MkdirAll(outPath, 0777)
	for _, art := range artists {

		artPath := outPath + "/" + art.Name()
		os.Mkdir(artPath, 0777)
		albumCount += len(art.albums)
		for _, alb := range art.albums {
			albPath := artPath + "/(" + alb.year + ") - " + alb.Name()
			os.Mkdir(albPath, 0777)
			//move the files
			if !dummy {
				files := getFiles(alb.path)
				fileCount += len(files)
				for _, v := range files {
					if v.IsDir() { //there can be nested albums
						continue
					}
					err := os.Rename(alb.path+"/"+v.Name(), albPath+"/"+v.Name())
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	fmt.Println(len(artists), "artists,", albumCount, "albums,", fileCount, "files")
}
