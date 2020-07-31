package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

//TODO: what to do about moana, Dr. Dre, starbound ost
//TODO: if an album has multiple AlbumArtist in one folder, complain
//TODO: how to create directory AC/DC
//TODO: handle multiple CD albums

var path = "/Users/cyrushanlon/Documents/Music"
var outPath = "/Users/cyrushanlon/Documents/Out"
var artists map[string]*artist

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

func main() {

	artists = make(map[string]*artist)

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			//get the extension of the file
			fileName := info.Name()
			ext := ""
			for i := len(fileName) - 1; i >= 0; i-- {
				if fileName[i] == '.' {
					break
				}
				ext = string(fileName[i]) + ext
			}

			//not a music file so we dont care about it
			if strings.ToLower(ext) != "mp3" && strings.ToLower(ext) != "mp4" && strings.ToLower(ext) != "flac" && strings.ToLower(ext) != "ogg" {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			m, err := tag.ReadFrom(f)
			if err != nil {
				log.Println("missing tags:", path)
				return nil
			}

			artistName := strings.TrimSpace(m.AlbumArtist())
			if artistName == "" {
				artistName = strings.TrimSpace(m.Artist())
				if artistName == "" {
					log.Println("missing name:", path)
					return nil
				}
			}

			albumName := strings.TrimSpace(m.Album())
			if albumName == "" {
				log.Println("missing album:", path)
				return nil
			}

			albumYear := m.Year()
			if albumYear == 0 {
				log.Println("missing year:", path)
				return nil
			}

			//get or create current artist/album
			var art *artist
			var alb *album
			var ok bool

			artistNameLower := strings.ToLower(artistName)
			albumNameLower := strings.ToLower(albumName)

			if art, ok = artists[artistNameLower]; !ok {
				art = &artist{
					albums:     make(map[string]*album),
					appearedAs: make(map[string]int),
				}
				artists[artistNameLower] = art
			}

			if alb, ok = art.albums[albumNameLower]; !ok {
				alb = &album{
					appearedAs: make(map[string]int),
					path:       path,
					year:       strconv.Itoa(albumYear),
				}
				art.albums[albumNameLower] = alb
			}

			art.appearedAs[artistName]++
			alb.appearedAs[albumName]++

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	//sort the keys for display only
	var keys []string
	for k := range artists {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, v := range keys {
		for _, j := range artists[v].albums {
			fmt.Println(artists[v].Name(), "-", j.Name())
		}
	}

	// now that we have the album/artist structure and info, we build the corresponding output

	os.MkdirAll(outPath, 0777)
	for _, art := range artists {
		artPath := outPath + "/" + art.Name()
		os.Mkdir(artPath, 0777)
		for _, alb := range art.albums {
			albPath := artPath + "/(" + alb.year + ") - " + alb.Name()
			os.Mkdir(albPath, 0777)
		}
	}

	fmt.Println(len(artists), "artists")
}
