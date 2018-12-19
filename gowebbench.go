package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// import "github.com/pkg/profile"

type bench struct {
	allrequests int
	allbytes    int
	successed   int
	failed      int
}

type benchitem struct {
	success  bool
	doclenth int
}

//GET或者POST
var method string

//执行总时间
var alltimes int

//并发实例数
var clients int

//总执行次数，此参数优先级高于alltimes 当allrequests不为零时 则按照总执行次数原则
var allrequests int

//POST提交的参数
var postdata string

//post数据保存的文件路径，优先级高于postdata，格式同postdata
var postdatafile string

//POST提交的文件路径
var postfile string

//POST提交的文件对应的字段名
var postfilename string

//提交的cookies组合，多组cookie用|分开
var cookies string

//提交的header头部组合，多个header用|分开
var headers string

//cookie保存的文件路径，优先级高于cookies,格式同cookies
var cookiefile string

//访问的URL地址
var url string

//当前版本号
const ver = "0.2"

var requestlist = make(map[int](*http.Request))
var mybench = make(chan benchitem)
var isstop = false
var wg sync.WaitGroup
var mux sync.Mutex
var benchs = bench{allrequests: 0, allbytes: 0, successed: 0, failed: 0}

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()

	if !initArgs() {
		return
	}

	maxProcs := runtime.NumCPU() //获取cpu个数
	runtime.GOMAXPROCS(maxProcs) //限制同时运行的goroutines数量

	cookielist := strings.Split(cookies, "|")
	for cindex, thecookie := range cookielist {
		req, err := buildRequest(url, method, postdata, postfile, postfilename, thecookie, headers)
		if err != nil {
			continue
		}
		requestlist[cindex] = req
	}
	fmt.Println("GoWebbench - Simple Web Benchmark using Go", ver)
	fmt.Println("Copyright (c) ebh.net 2011-2016.")
	fmt.Println("")
	fmt.Println("Benchmarking:", method, url)
	fmt.Println(clients, "clients, running", alltimes, "sec.")
	fmt.Println("")
	if len(requestlist) <= 0 {
		fmt.Println("request build error")
		return
	}

	// go func() {
	// 	wg.Wait()
	// 	fmt.Println("wait")
	// 	close(mybench)
	// }()
	if alltimes > 0 {
		time.AfterFunc(time.Duration(alltimes)*time.Second, endBench) //定时结束
	}
	var begintime = time.Now()
	for i := 1; i <= clients; i++ {
		// fmt.Println("clients", clients, " cur:", i)
		wg.Add(1)
		// fmt.Println("wg.add")
		// time.Sleep(time.Duration(10))
		go doBench(i)
	}
	go func() {
		for ben := range mybench {
			if ben.success {
				benchs.successed++
				benchs.allrequests++
				benchs.allbytes += ben.doclenth
			} else if ben.doclenth != -1 {
				benchs.failed++
				benchs.allrequests++
			}
			if allrequests > 0 && benchs.allrequests >= allrequests {
				endBench()

				// fmt.Println("do break")
				break
			}
		}

	}()
	// go func() {
	// fmt.Println("before wait")
	wg.Wait()
	// fmt.Println("wait")
	close(mybench)
	time.Sleep(1)
	// }()
	var endtime = time.Now()
	var duratime = endtime.Sub(begintime)
	var allseconds = int(duratime.Seconds())
	if allseconds == 0 {
		allseconds = 1
	}
	speed := int(benchs.successed / allseconds * 60)
	persecond := int(benchs.successed / allseconds)
	avgbytes := int(benchs.allbytes / allseconds)
	fmt.Println("Speed=", speed, " pages/min, ", avgbytes, " bytes/sec.")
	fmt.Println("Request per second:", persecond)
	fmt.Println("Requests: ", benchs.successed, " succeed, ", benchs.failed, " failed.")
}
func endBench() {
	mux.Lock()
	isstop = true
	mux.Unlock()
}
func doBench(index int) {
	defer wg.Done()
	// defer fmt.Println("do defer")
	// fmt.Println(1)
	ccount := len(requestlist)
	curindex := index % ccount
	var req *http.Request
	ok := true
	var thecookie string
	if postfile != "" || postdata != "" {
		cookielist := strings.Split(cookies, "|")
		thecookie = cookielist[curindex]
	} else {
		req, ok = requestlist[curindex]
	}
	if !ok {
		fmt.Println("build request err")
		return
	}
	// fmt.Println(2)
	// fmt.Println("dobench", index)
	// client := &http.Client{Timeout: 15 * time.Second}
	var trans http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: trans,
	}
	// fmt.Println(req)
	// fmt.Println(3)
	// ccc := 0
	for !isstop {
		// fmt.Println("allrequests", allrequests, "benchs.allrequests", benchs.allrequests)
		if allrequests > 0 && benchs.allrequests >= allrequests {
			fmt.Println("stop ", "benchs.allrequests:", benchs.allrequests, "allrequests:", allrequests)
			endBench()
			fmt.Println("stop endbench")
			item := benchitem{success: false, doclenth: -1}
			mybench <- item
			fmt.Println("beforem break end after <- ")
			// break
		}
		// ccc += 1
		// fmt.Println("before ccc:", ccc)
		if postfile != "" || postdata != "" { //如果有文件上传，则每次执行协程都要重新生成request
			var err error
			req, err = buildRequest(url, method, postdata, postfile, postfilename, thecookie, headers)
			if err != nil {
				ok = false
				fmt.Println("build request err")
				continue
			}
		}
		resp, err := client.Do(req)
		item := benchitem{}
		if err != nil || resp == nil || resp.Body == nil {
			item.success = false
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("is nil")
		} else {
			// defer resp.Body.Close()
			doclength, err := io.Copy(ioutil.Discard, resp.Body)
			// doclength, err := io.Copy(os.Stdout, resp.Body)
			if err != nil {
				fmt.Println("read err")
			}
			item.success = true
			item.doclenth = int(doclength)
			resp.Body.Close()
		}
		// fmt.Println("mid ccc:", ccc)
		mybench <- item
		// fmt.Printf("do %d\r", ccc)
	}
	// fmt.Println("before wg.done")
	// wg.Done()
}

/**
* 初始化参数
 */
func initArgs() bool {
	checkResult := true
	flag.StringVar(&url, "url", "", "bench url")
	flag.StringVar(&method, "m", "", "get or method default get")
	flag.IntVar(&alltimes, "t", 0, "bench time,if arg n set,it will be igorn")
	// flag.IntVar(&allrequests, "n", 0, "number request to bench")
	flag.IntVar(&clients, "c", 0, "clients")
	flag.StringVar(&headers, "h", "", "more headers split by | ")
	flag.StringVar(&postdata, "d", "", "post filed and value")
	flag.StringVar(&postdatafile, "dfile", "", "post data file path")
	flag.StringVar(&postfile, "f", "", "post filedata path")
	flag.StringVar(&postfilename, "fname", "", "post filedata field name")
	flag.StringVar(&cookies, "cookies", "", "more cookies split by | ")
	flag.StringVar(&cookiefile, "cfile", "", "cookies save file path")
	flag.Parse()
	allrequests = 0 //暂时屏蔽此参数
	if url == "" && flag.NArg() > 0 {
		url = flag.Arg(0)
	}
	if url == "" {
		fmt.Println("url 参数不能为空")
		checkResult = false
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Println("url 参数非有效的http格式")
		checkResult = false
	}
	method = strings.ToUpper(method)
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "POST" { //当前只支持get和post请求
		method = "GET"
	}
	if postdata != "" || postfile != "" || postdatafile != "" {
		method = "POST"
	}
	if postfile != "" && postfilename == "" {
		postfilename = "file"
	}
	if alltimes <= 0 {
		alltimes = 30
	}
	if allrequests > 0 { //n优先级高于n
		alltimes = 0
	}
	if clients <= 0 {
		clients = 10
	}
	//do postdatafile
	if postdatafile != "" {
		pfile, perr := ioutil.ReadFile(postdatafile)
		if perr != nil {
			checkResult = false
			fmt.Println(perr)
		} else {
			postdata = string(pfile)
		}
	}
	//do cookiefile
	if cookiefile != "" {
		cfile, cerr := ioutil.ReadFile(cookiefile)
		if cerr != nil {
			checkResult = false
			fmt.Println(cerr)
		} else {
			cookies = string(cfile)
		}
	}
	if !checkResult {
		flag.Usage()
	}
	return checkResult
}

/*
* 生成请求对象
 */
func buildRequest(burl, bmethod, bpostdata, bpostfile, bpostfilename, bcookie string, bheader string) (*http.Request, error) {
	contentType := "text/html;charset=UTF-8"
	var body *bytes.Buffer
	if bpostdata != "" && bpostfile != "" {
		body = &bytes.Buffer{}
		// fmt.Println(bpostdata)
		multiwriter := multipart.NewWriter(body)

		if bpostdata != "" { //处理post字段请求
			postfields := strings.Split(bpostdata, "&")
			for _, postfield := range postfields {
				fieldkeyvalue := strings.Split(postfield, "=")
				if len(fieldkeyvalue) < 2 {
					continue
				}
				fieldname := fieldkeyvalue[0]
				// fmt.Println("fieldname", fieldname)
				fieldvalue := strings.Join(fieldkeyvalue[1:], "=")
				// fmt.Println("fieldvalue", fieldvalue)
				multiwriter.WriteField(fieldname, fieldvalue)
			}
		}
		if bpostfile != "" { //处理文件上传
			file, fileerr := multiwriter.CreateFormFile(bpostfilename, bpostfile)
			if fileerr != nil {
				return nil, fileerr
			}
			pfile, perr := os.Open(bpostfile)
			if perr != nil {
				fmt.Println(perr)
				return nil, perr
			}

			io.Copy(file, pfile)
			defer pfile.Close()
		}
		contentType = multiwriter.FormDataContentType()
		multiwriter.Close()
	} else if postdata != "" {
		body = bytes.NewBufferString(postdata)
		contentType = "application/x-www-form-urlencoded"
	}
	if body == nil {
		body = &bytes.Buffer{}
	}
	req, err := http.NewRequest(bmethod, burl, body) //此方法必须传POST或者GET,如果为GET 则其他postfile等数据不会提交
	if err != nil {
		return nil, err
	}
	if bcookie != "" { //处理cookie
		//过滤cookie特殊字符
		bcookie = strings.Replace(bcookie, ";", "&", -1)
		bcookie = strings.Replace(bcookie, "\r", "", -1)
		bcookie = strings.Replace(bcookie, "\n", "", -1)
		cookiefields := strings.Split(bcookie, "&")
		for _, cookiefield := range cookiefields {
			cfieldvalue := strings.Split(cookiefield, "=")
			if len(cfieldvalue) < 2 {
				continue
			}
			cookiename := cfieldvalue[0]
			cookiename = strings.Trim(cookiename, " ")
			cookievalue := strings.Join(cfieldvalue[1:], "=")
			cookievalue = strings.Trim(cookievalue, " ")
			cook := &http.Cookie{Name: cookiename, Value: cookievalue}
			req.AddCookie(cook)
		}
	}
	req.Header.Set("Content-Type", contentType)
	if bheader != "" { //处理头部header
		headerfields := strings.Split(bheader, "&")
		for _, headerfield := range headerfields {
			cfieldvalue := strings.Split(headerfield, "=")
			if len(cfieldvalue) < 2 {
				continue
			}
			headername := cfieldvalue[0]
			headername = strings.Trim(headername, " ")
			headervalue := strings.Join(cfieldvalue[1:], "=")
			headervalue = strings.Trim(headervalue, " ")
			req.Header.Add(headername, headervalue)
		}
	}
	return req, nil
}
