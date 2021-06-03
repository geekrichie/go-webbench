package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var  timerexpired atomic.Value
var speed  = 0
var bytes  = 0

var http_version = 1/*0 - http/0.9 1- http/1.0 2 - http/1.1 */
/* Allow GET, HEAD, OPTIONS, TRACE*/
const  (
	METHOD_GET = iota
	METHOD_HEAD
	METHOD_OPTIONS
	METHOD_TRACE
)
const PROGRAM_VERSION = "1.5"

var method = METHOD_GET
var proxyport=80
var proxyhost string = ""


var mypipe [2]int
var host string
var request string


var force int
var force_reload int
var benchtime int
var help  string
var http10 int
var clients int
var version string
var proxy string

func init() {
	flag.IntVar(&force,"force", 0, "Don't wait for reply from server.")
	flag.IntVar(&force_reload,"reload",0, "Send reload request - Pragma: no-cache.")
	flag.IntVar(&benchtime, "time", 5,"Run benchmark for <sec> seconds. Default 30.")
	flag.StringVar(&help,"help","?","This information.")
	flag.IntVar(&http10,"http", 1,"http version")
	flag.IntVar(&clients,"clients",1,"Run <n> HTTP clients at once. Default one.")
	flag.StringVar(&version,"version","1.5","Display program version.")
	flag.StringVar(&proxy,"proxy","","Use proxy server for request")
}

func usage() {
	fmt.Fprint(os.Stderr,
		"webbench [option]... URL\n" +
			"  -f|--force               Don't wait for reply from server.\n" +
			"  -r|--reload              Send reload request - Pragma: no-cache.\n" +
			"  -t|--time <sec>          Run benchmark for <sec> seconds. Default 30.\n" +
			"  -p|--proxy <server:port> Use proxy server for request.\n" +
			"  -c|--clients <n>         Run <n> HTTP clients at once. Default one.\n" +
			"  -9|--http09              Use HTTP/0.9 style requests.\n" +
			"  -1|--http10              Use HTTP/1.0 protocol.\n" +
			"  -2|--http11              Use HTTP/1.1 protocol.\n" +
			"  --get                    Use GET request method.\n" +
			"  --head                   Use HEAD request method.\n" +
			"  --options                Use OPTIONS request method.\n" +
			"  --trace                  Use TRACE request method.\n" +
			"  -?|-h|--help             This information.\n" +
			"  -V|--version             Display program version.\n")

}

func main() {
	if len(os.Args) ==1 {
		usage()
		return
	}
	//解析参数
	flag.Parse()
	if proxy != "" {

		tmp := strings.LastIndex(proxy, ":")

		if tmp == -1 {
			usage()
			return
		} else if tmp == 0 {
			fmt.Fprintf(os.Stderr, "Error in option --proxy %s: Missing hostname.\\n", proxy)
			return
		} else if tmp == len(proxy) {
			fmt.Fprintf(os.Stderr, "Error in option --proxy %s Port number is missing.\n", proxy)
			return
		}

		proxyport, _ = strconv.Atoi(proxy[tmp+1:])
		proxyhost = proxy[:tmp]
	}

	fmt.Fprint(os.Stderr,"Webbench - Simple Web Benchmark "+PROGRAM_VERSION +"\n"+
		"Copyright (c) Radim Kolar 1997-2004, GPL Open Source Software.\n")

	if flag.NArg() <= 0 {
		fmt.Fprint(os.Stderr, "缺少URL参数")
	}
	build_request(flag.Arg(0))

	log.Println("Running info: ")

	if clients  == 1 {
		fmt.Print(" 1 client")
	}else {
		fmt.Printf("%d clients", clients)
	}

	fmt.Printf(", running %d sec", benchtime)

	if force  == 1 {
		fmt.Print(", early socket close")
	}

	if proxyhost != "" {
		fmt.Printf(", via proxy server %s:%d", proxyhost, proxyport)
	}

	if force_reload == 1 {
		fmt.Print(", forcing reload")
	}

	fmt.Println()

	bench()
	return
}

func build_request(url string) {

	if force_reload  == 1&& proxyhost != "" && http10 < 1  {
		http10 = 1
	}
	if method==METHOD_HEAD && http10<1{
		http10=1
	}
	if method==METHOD_OPTIONS && http10<2{
		http10=2
	}
	if method==METHOD_TRACE && http10<2 {
		http10=2
	}

	switch method {
	default:
	case METHOD_GET:
		request = "GET"
	case METHOD_HEAD:
		request = "HEAD"
	case METHOD_OPTIONS:
		request = "OPTIONS"
	case METHOD_TRACE:
		request = "TRACE"
	}
	request = request + " "

	if strings.Index(url,"://") == -1 {
		fmt.Fprintf(os.Stderr, "\n%s: is not a valid URL. \n", url)
		os.Exit(2)
	}

	if len(url) > 1500 {
		fmt.Fprintln(os.Stderr, "URL is too long.\n", url)
		os.Exit(2)
	}

	if strings.HasPrefix(url,"http://") == false {
		fmt.Fprintln(os.Stderr, "\nOnly HTTP protocol is directly supported, set --proxy for others.\n", url)
		os.Exit(2)
	}

	i := strings.Index(url, "://") + 3
	if strings.Index(url[i:], "/") == -1 {
		fmt.Fprintln(os.Stderr, "\nInvalid URL syntax - hostname don't ends with '/'.\n", url)
		os.Exit(2)
	}

	if proxyhost == "" {
		/* get port from hostname */
		if strings.Index(url[i:], ":") != -1 &&
			strings.Index(url[i:], ":") < strings.Index(url[i:], "/") {
			host = url[i:i+strings.Index(url[i:], ":")]
			proxyport,_ = strconv.Atoi(url[i+1+strings.Index(url[i:], ":") : i+strings.Index(url[i:], "/")])
			if proxyport  == 0 {
				proxyport  = 80
			}
		}else {
			host = url[i:i+strings.Index(url[i:], "/")]
		}
		//fmt.Println(url[i+1+strings.Index(url[i:], ":") : i+strings.Index(url[i:], "/")])
		request = request + url[i+strings.Index(url[i:], "/"):]
	}else {
		request = request + url
	}

	if http10 == 1 {
		request = request + " HTTP/1.0"
	}else if http10 == 2 {
		request = request + " HTTP/1.1"
	}
	request = request  + "\r\n"

	if http10 > 0 {
		request = request + "User-Agent: WebBench " + PROGRAM_VERSION + "\r\n"
	}
	if proxyhost == "" && http10 > 0 {
		request  += "Host: " + host +"\r\n"
	}

	if force_reload == 1 && proxyhost != "" {
		request = request + "Pragma: no-cache\r\n"
	}

	if http10 > 1 {
		request = request + "Connection: close\r\n"
	}

	if http10 > 0 {
		request =  request + "\r\n"
	}

	log.Printf("\nRequest:\n%s\n", request)
}
type comm struct {
	speed  int
	failed int
	byte   int
}

func bench(){
	var (
		c  net.Conn
		err error
		cm comm
	)
	chs  := make([]chan comm, clients)
	for i:=0;i  < clients;i++ {
		chs[i] = make(chan comm)
	}
	/* check avaibility of target server */
	if proxyhost != "" {
		c , err = Socket(proxyhost, proxyport)
	}else {
		c, err = Socket(host, proxyport)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr,"\nConnect to server failed. Aborting benchmark." )
	}
	c.Close()
	//启动协程并发访问请求
	for i := 0 ; i < clients; i++ {
		if proxyhost == "" {
			go benchscore(host, proxyport, request,chs[i])
		}else {
			go benchscore(proxyhost, proxyport, request, chs[i])
		}
	}


	for i := 0 ; i < clients; i++ {
		data := <-chs[i]
		cm.speed += data.speed
		cm.failed += data.failed
		cm.byte += data.byte
	}
	fmt.Printf("\nSpeed=%d pages/min, %d bytes/sec.\nRequests: %d susceed, %d failed.\n",
		int(float64(cm.speed + cm.failed)/(float64(benchtime)/60.0)),
		int(float64(cm.byte)/float64(benchtime)),
		cm.speed,
		cm.failed)

}

func benchscore(host string, port int, request string, ch chan comm) {
	var (
		c net.Conn
		err error
		n int
		d interface{}
		cm comm
		i int
	)
	buf := make([]byte, 1500)
	var timeexpried atomic.Value
	timeexpried.Store(0)
	ticker1 := time.NewTicker(time.Duration(benchtime)* time.Second)
	defer ticker1.Stop()
	go func (){
		fmt.Println("管理超时的协程已经启动")
		select {
		case <- ticker1.C :
			fmt.Println("已经超时")
			timeexpried.Store(1)
			return
		}
	}()
	for{
		if d = timeexpried.Load() ; d.(int)  == 1{
			if cm.failed > 0 {
				cm.failed--
			}
			ch <- cm
			fmt.Println("协程执行结束")
			return
		}
		c, err = Socket(host, port)
		if err != nil {
			cm.failed++
			continue
		}
		n, err = c.Write([]byte(request))
		if err != nil {
			cm.failed++
			continue
		}

		if  n < len(request) {
			cm.failed++
			continue
		}
		if force  == 0 {
			for {
				//这里需要对timerexpired做下处理,要不然go可能对timerexpired有优化
				if  d = timeexpried.Load() ; d.(int)  == 1 {
					break
				}
				i, err = c.Read(buf)
				if i<0 {
					break
				}else if i == 0 {
					break
				}else {
					cm.byte += i
				}
			}
		}
		err = c.Close()
		if err != nil {
			cm.failed ++
			continue
		}
		cm.speed ++
	}

}