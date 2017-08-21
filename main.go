package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"net/url"
	"crypto/tls"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(fmt.Sprintf("%+v", r))
	headerConnectionValue := r.Header.Get("Proxy-Connection")
	if headerConnectionValue == "" {
		w.WriteHeader(400)
		return
	}
	fmt.Println(r.URL.String())
	fmt.Println(r.URL.Path)
	if r.Method == "GET" {
		request, err := http.NewRequest(r.Method, r.URL.String(), nil)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		for k, v := range r.Header {
			if k == "Proxy-Connection" {
				k = "Connection"
			}
			request.Header.Set(k, v[0])
		}
		proxy, _ := url.Parse("http://124.238.235.135:81")
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		timeout := time.Duration(time.Duration(10) * time.Second)
		httpClient := &http.Client{Transport:tr, Timeout:timeout}
		fmt.Println(fmt.Sprintf("%+v", request))
		res, err := httpClient.Do(request)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		body, err := ioutil.ReadAll(res.Body)
		for k, v := range res.Header {
			w.Header().Set(k, v[0])
		}
		w.Write(body)
		w.WriteHeader(res.StatusCode)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}