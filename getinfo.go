package main

import (
	"bufio"
	"errors"
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
			if url != "" {
				fmt.Print(1, url)
				fmt.Println(getinfo(url))
			}
			if file != "" {
				f, err := os.Open(file)
				if err != nil {
					log.Println(err)
				}
				scanner := bufio.NewScanner(f)
				defer f.Close()

				record, err := os.Create("./response.txt")
				if err != nil {
					fmt.Println(err)
				}
				log.SetOutput(record)
				defer record.Close()
				log.SetFlags(0)

				log.Println("id", "url", "statusCode", "startTime", "endTime", "size")
				id := 0
				for scanner.Scan() {
					wg.Add(1)
					go addToFile(id, scanner.Text(), &wg)
					id++
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

// adds the getinfo information to a logfile for one url
func addToFile(id int, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	status, start, end, size, err := getinfo(url)
	if err != nil {
		return
	}
	log.Println(id, url, status, start, end, size)
}

// getinfo function, makes request to a url as a string, returns status, start time, end time, size
func getinfo(url string) (int, int64, int64, int64, error) {
	if len(url) < 8 || url[:8] != "https://" {
		url = "https://" + url
	}

	start := time.Now().UnixNano()
	resp, err := http.Get(url)
	end := time.Now().UnixNano()
	if err != nil  {
		return -1, start, end, 0, errors.New("error making request")
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
	return resp.StatusCode, start, end, size, nil
}
