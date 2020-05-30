package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

var iptable []string
var mutex sync.Mutex
var lastip string = ""

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
	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
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
	if len(ipSlice) > 300 {
		ipSlice = ipSlice[:300]
	}
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

//使用`http`代理访问一个网页
func GetHttp(httpUrl, proxy_addr string) ([]byte, error) {
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
		//fmt.Println(err.Error())
		return []byte{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		//fmt.Println("err statusCode:", res.StatusCode)
		return []byte{}, errors.New("retcode:" + string(res.StatusCode))
	}
	c, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err.Error())
		return []byte{}, err
	}
	if len(c) == 0 {
		return []byte{}, errors.New("not any response")
	}
	return c, nil
}

//并发的访问一个网页
func GetHttpSelect(httpUrl string) ([]byte, error) {
	type chs struct {
		ret []byte
		ip  string
		err error
	}

	//获取上一次访问的ip
	mutex.Lock()
	ip := lastip
	mutex.Unlock()

	if ip != "" {
		ch := make(chan chs)
		go func() {
			ret, err := GetHttp(httpUrl, ip)
			if err == nil {
				ch <- chs{ret, ip, nil}
			}

		}()
		select {
		case result := <-ch:
			if result.err != nil {
				mutex.Lock()
				lastip = ""
				mutex.Unlock()
				return []byte{}, result.err
			}
			return result.ret, nil
		case <-time.After(10 * time.Second):
			mutex.Lock()
			lastip = ""
			mutex.Unlock()
			return []byte{}, errors.New("time out err")
		}
	}

	//并发访问
	ch := make(chan chs)
	for i := 0; i < 15; i++ {
		go func() {
			mutex.Lock()
			i := rand.Intn(len(iptable))
			s := iptable[i]
			mutex.Unlock()
			ret, err := GetHttp(httpUrl, s)
			if err == nil {
				ch <- chs{ret, s, nil}
			}
		}()
	}
	select {
	case result := <-ch:
		if result.err == nil {
			mutex.Lock()
			lastip = result.ip
			mutex.Unlock()
		}
		return result.ret, nil
	case <-time.After(10 * time.Second):
		return []byte{}, errors.New("time out err")
	}
}

//从本地获取端口号
func GetPort() (string, error) {

	is_exist := func(path string) bool {
		_, err := os.Lstat(path)
		return !os.IsNotExist(err)
	}("config.txt")
	if !is_exist {
		ioutil.WriteFile("config.txt", []byte(`{"port":"5000"}`), 0777)
	}

	data, err := ioutil.ReadFile("config.txt")
	if err != nil {
		return "", err
	}
	type ConfigType struct {
		Port string
	}
	var s ConfigType
	err = json.Unmarshal([]byte(data), &s)
	if err != nil {
		return "", err
	}
	return s.Port, nil
}

//程序入口
func main() {
	//初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	fmt.Println("------欢迎使用ip代理获取工具------")
	fmt.Println("------Made by super1207------")
	fmt.Println("正在获取ip...")
	iptable = VIP(GetIp())
	if len(iptable) == 0 {
		log.Fatal("没有获取到任何ip")
	}
	fmt.Println("当前IP数量：", len(iptable))
	//每分钟刷新一次ip
	go func() {
		for {
			time.Sleep(60 * time.Second)
			iptable_t := VIP(GetIp())
			mutex.Lock()
			if len(iptable_t) != 0 {
				iptable = iptable_t
			}
			fmt.Println("当前IP数量：", len(iptable))
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

	//使用http代理访问一个网页
	http.HandleFunc("/get_url", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["url"]
		if !ok || len(keys) < 1 {
			log.Println("Url Param 'key' is missing")
			w.Write([]byte("no url input"))
		} else {
			mutex.Lock()
			ip := lastip
			mutex.Unlock()
			log.Printf("使用ip:'%s'访问网页：'%s'", ip, keys[0])
			cont := []byte{}
			var err error = nil
			for i := 0; i < 3; i++ {
				cont, err = GetHttpSelect(keys[0])
				if err == nil {
					break
				}
			}
			if err != nil {
				log.Printf("使用ip:'%s'访问网页：'%s'失败:'%s'", ip, keys[0], err.Error())
			}
			w.Header().Set("Content-Type", "text/html; charset=gbk")
			w.Write(cont)
		}
	})

	//切换ip
	http.HandleFunc("/change_ip", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		lastip = ""
		mutex.Unlock()
		w.Write([]byte("OK"))
	})

	//从本地文件获得端口号
	port, err := GetPort()
	if err != nil {
		log.Fatal(err.Error())
	}

	helpFun := func() {
		fmt.Printf("访问地址：http://localhost:%s/get_ip\n", port)
		fmt.Printf("访问地址：http://localhost:%s/get_all\n", port)
		fmt.Printf("访问地址：http://localhost:%s/change_ip\n", port)
		fmt.Printf("访问地址：http://localhost:%s/get_url?url=http://vip.stock.finance.sina.com.cn/corp/go.php/vCI_StockHolder/stockid/601778/displaytype/30.phtml\n", port)
	}

	//处理控制台输入
	go func() {
		for {
			in := bufio.NewReader(os.Stdin)
			str, _, err := in.ReadLine()
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			if string(str) == "help" {
				helpFun()
			}
		}
	}()

	//启动服务器
	fmt.Println("启动http seriver...")
	helpFun()
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

