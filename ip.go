package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var iptable []string
var mutex sync.Mutex

//判断代理是否可用，，超时10秒，参数proxy_addr格式：`http://180.97.33.144:81`
func VerifIp(proxy_addr string) bool {
	httpUrl := "http://www.baidu.com"
	proxy, err := url.Parse(proxy_addr)
	netTransport := &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * time.Duration(10),
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	res, err := httpClient.Get(httpUrl)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return false
	}
	return true
}

//从89ip获得代理ip包含代理ip的切片
func GetIp() (ipSlice []string) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Get("http://www.89ip.cn/tqdl.html?api=1&num=9999&port=&address=&isp=")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	c, _ := ioutil.ReadAll(resp.Body)
	comp := regexp.MustCompile("(\\d+\\.\\d+\\.\\d+\\.\\d+:\\d+)")
	submatchs := comp.FindAllStringSubmatch(string(c), -1)
	for _, submatch := range submatchs {
		ipSlice = append(ipSlice, "http://"+submatch[0])
	}
	return
}

//打乱切片,注意会打乱自身
func Random(strings []string) []string {
	for i := len(strings) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		strings[i], strings[num] = strings[num], strings[i]
	}
	return strings
}

//得到可用的ip
func VIP(ipSlice1 []string) []string {
	type result struct {
		string
		bool
	}
	ipSlice := Random(ipSlice1)
	ipSlice = ipSlice[:300]
	num := 0
	canuseNum := 0
	resultChannel := make(chan result)
	for i := 0; i < len(ipSlice); i++ {
		ip := ipSlice[i]
		go func() {
			is_ok := VerifIp(ip)
			resultChannel <- result{ip, is_ok}
		}()
	}
	var ret []string
	for i := 0; i < len(ipSlice); i++ {
		result := <-resultChannel
		num++
		if result.bool {
			canuseNum++
			ret = append(ret, result.string)
		}
	}
	return ret
}

//程序入口
func main() {
	//初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	fmt.Println("------欢迎使用ip代理获取工具------")
	fmt.Println("------Made by super1207------")
	fmt.Println("正在获取ip...")
	iptable = VIP(GetIp())
	fmt.Println("当前IP数量：", len(iptable))

	//每分钟刷新一次ip
	go func() {
		for i := 0; i != 1; i = 0 {
			time.Sleep(60 * time.Second)
			iptable_t := VIP(GetIp())
			fmt.Println("当前IP数量：", len(iptable_t))
			mutex.Lock()
			iptable = iptable_t
			mutex.Unlock()

		}

	}()

	//随机获得一个ip
	http.HandleFunc("/get_ip", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		i := rand.Intn(len(iptable))
		s := iptable[i]
		mutex.Unlock()
		w.Header().Set("content-type", "application/json")
		w.Write([]byte("[\"" + s + "\"]"))
	})

	//获得所有ip
	http.HandleFunc("/get_all", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		data, _ := json.Marshal(iptable)
		mutex.Unlock()
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(data))
	})

	//启动服务器
	fmt.Println("启动http seriver...")
	fmt.Println("访问地址：http://localhost:5000/get_ip")
	fmt.Println("访问地址：http://localhost:5000/get_all")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
