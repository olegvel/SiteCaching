package main

import (
	"bytes"
	"fmt"
	"github.com/golang/groupcache/lru"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var url = "http://108.61.245.170"
var headerPrefix = "header_"

func main() {
	var cache = &lru.Cache{}
	Caching(url, cache)
	Caching(url+"/image.jpg", cache)

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for t := range ticker.C {
			Caching(url, cache)
			Caching(url+"/image.jpg", cache)
			fmt.Println("Updated at", t)
		}
	}()

	http.Handle("/", SiteHandler(url, cache))
	http.Handle("/image.jpg", SiteHandler(url+"/image.jpg", cache))
	if err := http.ListenAndServe(":1080", nil); err != nil {
		log.Fatal(err)
	}
}

func SiteHandler(url string, cache *lru.Cache) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp, ok := cache.Get(url)
		if !ok {
			log.Fatal("error getting")
		}

		header, ok := cache.Get(headerPrefix + url)
		if !ok {
			log.Fatal("error getting")
		}

		for k, v := range header.(http.Header) {
			fmt.Printf("Header field %q, Value %q\n", k, v)
			writer.Header().Set(k, v[0])
		}

		response := bytes.NewReader(resp.([]byte))

		_, err := io.Copy(writer, response)
		if err != nil {
			log.Fatal(err)
		}

	}
}

func Caching(url string, cache *lru.Cache) error {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	byteBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cache.Add(headerPrefix+url, resp.Header)
	cache.Add(url, byteBody)

	return nil
}
