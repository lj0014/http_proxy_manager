package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"net/url"
	"crypto/tls"
	"sync"
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
		res := getByProxy(request)
		if res == nil {
			w.WriteHeader(500)
			return
		}
		fmt.Println(fmt.Sprintf("%+v", res))

		for k, v := range res.Header {
			w.Header().Set(k, v[0])
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
		}
		w.Write(body)

		w.WriteHeader(res.StatusCode)
	}
}

func getByProxy(request *http.Request) *http.Response {
	for i:=0; i<3; i++ {
		proxyURL := getProxy()
		proxy, _ := url.Parse(proxyURL)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		timeout := time.Duration(time.Duration(10) * time.Second)
		httpClient := &http.Client{Transport:tr, Timeout:timeout}
		fmt.Println(fmt.Sprintf("%+v", request))
		fmt.Println("1111111111111111111112222222222222222222222")
		res, err := httpClient.Do(request)
		if err != nil {
			fmt.Println(err)
			//w.WriteHeader(500)
			putProxy(proxyURL, false)
			continue
		}
		fmt.Println("111111111111111111111")
		putProxy(proxyURL, true)
		return res
	}
	return nil
}

type ProxyData struct {
	URL string
	LastErrorCount int
	ErrorCount int
	SuccessCount int
	OK bool
	LastOK int64
}

var proxyLock *sync.Mutex
var proxyArray []ProxyData
func loadProxy() {
	proxyLock.Lock()
	defer proxyLock.Unlock()
	addProxy("http://124.238.235.135:81")
	addProxy("http://58.100.105.28:8888")
}

func dumpProxy() {
	proxyLock.Lock()
	defer proxyLock.Unlock()
	fmt.Println(fmt.Sprintf("%+v", proxyArray))
}

func addProxy(proxyURL string) {
	for i, _ := range proxyArray {
		if proxyArray[i].URL == proxyURL {
			return
		}
	}
	var one ProxyData
	one.URL = proxyURL
	one.OK = true
	proxyArray = append(proxyArray, one)
}

func getProxy() string {
	proxyLock.Lock()
	fmt.Println("11111111111213123123")
	defer proxyLock.Unlock()
	var one ProxyData
	for {
		one = proxyArray[0]
		if one.OK {
			proxyArray = append(proxyArray[1:], one)
			break
		}
	}
	return one.URL
}

func putProxy(proxyURL string, isOK bool) {
	proxyLock.Lock()
	defer proxyLock.Unlock()
	for i, _ := range proxyArray {
		if proxyArray[i].URL == proxyURL {
			if isOK {
				proxyArray[i].SuccessCount = proxyArray[i].SuccessCount + 1
				proxyArray[i].LastOK = time.Now().Unix()
			} else {
				proxyArray[i].ErrorCount = proxyArray[i].ErrorCount + 1
				proxyArray[i].LastErrorCount = proxyArray[i].LastErrorCount + 1
				if proxyArray[i].LastErrorCount >= 3 {
					proxyArray[i].OK = false
				}
			}
			break
		}
	}
}

func autoLoad() {
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			loadProxy()
			dumpProxy()
		}
	}()
}

func main() {
	proxyLock = new(sync.Mutex)
	loadProxy()
	autoLoad()
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}