package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"landzero.net/x/encoding/yaml"
)

var datePattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

var configFileGlob = "/etc/dellog.d/*"

// Config config struct
type Config struct {
	File   string `yaml:"file"`
	Keep   int    `yaml:"keep"`
	Enable bool   `yaml:"enable"`
}

// DryRun flag
var DryRun bool

func main() {
	// options
	flag.BoolVar(&DryRun, "dry-run", false, "set dry-run flag, no file will be deleted")
	flag.Parse()

	var err error
	// current time
	n := time.Now()
	now := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
	// search config files
	var configFiles []string
	log.Println("scan  :", configFileGlob)
	if configFiles, err = filepath.Glob(configFileGlob); err != nil {
		log.Println("error :", err)
		return
	}
	log.Println("found :", len(configFiles), "config file(s)")
	// search log files
	for _, configFile := range configFiles {
		log.Println("------------------------------------")
		log.Println("load  :", configFile)
		// read config file
		var buf []byte
		if buf, err = ioutil.ReadFile(configFile); err != nil {
			log.Println("error : failed to read", configFile, err)
			return
		}
		// decode file
		var config Config
		if err = yaml.Unmarshal(buf, &config); err != nil {
			log.Println("error : failed to decode", configFile, err)
			return
		}
		if !config.Enable {
			continue
		}
		if config.Keep < 1 {
			log.Println("error : invalid field 'keep'")
			continue
		}
		config.File = strings.TrimSpace(config.File)
		if len(config.File) == 0 {
			log.Println("error : empty field 'file'")
			continue
		}
		// search files
		var files []string
		if files, err = filepath.Glob(config.File); err != nil {
			log.Println("error : invalid filed 'file'")
			continue
		}
		log.Println("scan  :", config.File)
		log.Println("keep  :", config.Keep, "day(s)")
		log.Println("------------------------------------")
		// check files
		for _, file := range files {
			// check dir
			var st os.FileInfo
			if st, err = os.Stat(file); err != nil || st.IsDir() {
				continue
			}
			// check filename
			var match []string
			if match = datePattern.FindStringSubmatch(filepath.Base(file)); len(match) != 1 {
				continue
			}
			// parse time
			var t time.Time
			if t, err = time.Parse("2006-01-02", match[0]); err != nil {
				continue
			}
			// compare time
			if now.Sub(t)/(time.Hour*24) > time.Duration(config.Keep) {
				if DryRun {
					log.Println("drydel:", file)
				} else {
					log.Println("delete:", file)
					os.Remove(file)
				}
			} else {
				log.Println("skip  :", file)
			}
		}
	}
}
