package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

// adds the getinfo information to a logfile for one url using a waitgroup for concurrency
func getInfo2(id string, url string, wg *sync.WaitGroup) {
	getinfo(id, url)
	wg.Done()
}

// adds the getinfo information to a logfile for one url - id, url, status code, start and end time, size
func getinfo(id string, url string) {
	if len(url) < 8 || url[:8] != "https://" {
		url = "https://" + url
	}

	start := time.Now().UnixNano()
	resp, err := http.Get(url)
	end := time.Now().UnixNano()
	if err != nil  {
		log.Println(id, url[8:], nil, start, end, nil, err.Error())
		return
	}
	defer resp.Body.Close()

	// resp.ContentLength doesn't seem to work very well
	size := resp.ContentLength
	if size == -1 {
		// this counts the bytes of of the response body, if ContentLength fails
		content, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			size = int64(len(content))
		}
	}
	log.Println(id, url[8:], resp.StatusCode, start, end, size, nil)
}

// I wasn't really sure what to do for a request ID, it seems like
// using a random string is pretty common though? Hope that's sufficient
// random_id creates a random 8-character string for an id
func random_id() (id string, err error) {
	b := make([]byte, 4)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	id = fmt.Sprintf("%x", b)
	return
}

func main() {
	var url string
	var file string
	var wg sync.WaitGroup

	app := &cli.App{
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name: "url",
				Usage: "url to be tested",
				Destination: &url,
			},
			&cli.StringFlag{
				Name: "file",
				Usage: "file to be tested",
				Destination: &file,
			},
		},
		Action: func(c *cli.Context) error {
			// sets up a log file
			if c.NArg() > 0 {
				if c.Args().Get(0) == "newlog" {
					record, err := os.Create("response.log")
					if err != nil {
						fmt.Println(err)
						return err
					}
					log.SetOutput(record)
					log.SetFlags(0)
					log.Println( "log_date", "log_time","id", "url", "status", "start", "end", "size", "err")
					defer record.Close()
					return nil
				}
			}

			record, err := os.OpenFile("response.log",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
			}
			log.SetOutput(record)
			defer record.Close()

			// handles the url flag
			if c.FlagNames()[0] == "url" {
				id, _ := random_id()
				getinfo(id, url)
			}
			// handles the file flag
			if c.FlagNames()[0] == "file" {
				f, err := os.Open(file)
				if err != nil {
					log.Println(err)
					return err
				}
				scanner := bufio.NewScanner(f)
				defer f.Close()

				for scanner.Scan() {
					id, _ := random_id()
					wg.Add(1)
					go getInfo2(id, scanner.Text(), &wg)
				}
				wg.Wait()
			}

			return nil
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}