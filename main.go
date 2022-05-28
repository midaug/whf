package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// "reflect"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/robertkrimen/otto"
)

var tp string           //模板路径
var jsDateFormat string //js定义的默认方法
//初始化加载
func runInit() {
	jsDateFormat = `function dateFormat(date, fmt) {
        if (!fmt){
            fmt = "YYYY-mm-dd HH:MM:SS.sss";
        }
        var ret;
        var opt = {
            "Y+": date.getFullYear().toString(),        // 年
            "m+": (date.getMonth() + 1).toString(),     // 月
            "d+": date.getDate().toString(),            // 日
            "H+": date.getHours().toString(),           // 时
            "M+": date.getMinutes().toString(),         // 分
            "S+": date.getSeconds().toString(),         // 秒
            "s+": date.getMilliseconds().toString()     // 毫秒
            // 有其他格式化字符需求可以继续添加，必须转化成字符串
        };
        function padValue(value) {
            return (value < 10) ? "0" + value : value;
        }
        for (var k in opt) {
            ret = new RegExp("(" + k + ")").exec(fmt);
            if (ret) {
                fmt = fmt.replace(ret[1], (ret[1].length == 1) ? (opt[k]) : padValue(parseInt(opt[k])));
            };
        };
        return fmt;
    }
    `
}

//校验路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//http返回封装
func requestReturn(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if code != 200 {
		log.Println("Err: requestReturn ", msg)
	}
	fmt.Fprintf(w, msg)
}

//otto中获取变量值
func getValueToString(vm *otto.Otto, name string) string {
	if value, err := vm.Get(name); err == nil {
		if valueStr, err := value.ToString(); err == nil {
			return valueStr
		}
	}
	return ""
}
//发送post请求
func httpPostJson(u string, contentType string, body string) {
	bodyBytes := []byte(body)
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	statuscode := resp.StatusCode
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Println("info: httpPostJson > code=", statuscode, " body=", string(respBody))
}

//http路由方法 /send
func send(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // 解析参数，默认是不会解析的
	log.Println("info: =======================================================")
	log.Println("info: Request Form => ", r.Form) // 这些信息是输出到服务器端的打印信息
	u := strings.Join(r.Form["u"], "")
	tn := strings.Join(r.Form["tn"], "")
	title := strings.Join(r.Form["title"], "")
	if len(u) == 0 {
		requestReturn(w, 400, "Err! The request parameter u is empty")
		return
	}
	if len(tn) == 0 {
		requestReturn(w, 400, "Err! The request parameter tn is empty")
		return
	}

	//先读入js模板文件内容
	jsFilePath := tp + "/" + tn + ".js"
	jsBytes, err := ioutil.ReadFile(jsFilePath)
	if err != nil {
		requestReturn(w, 400, "Err! tn file is not found > "+jsFilePath)
		return
	}

	body := make([]byte, r.ContentLength)
	r.Body.Read(body)

	vm := otto.New()
	vm.Set("bodyStr", string(body[:]))
	vm.Set("title", title)
	_, err = vm.Run(jsDateFormat + string(jsBytes))
	if err != nil {
		log.Println("Err: run js error > ", err)
		requestReturn(w, 500, "run js error!")
		return
	}

	var sendBodyStr = getValueToString(vm, "body")
	var contentType = getValueToString(vm, "contentType")
	if contentType == "undefined" || len(contentType) <= 0 {
		contentType = "application/json"
	}
	log.Println("info: js run sendBodyStr > ", sendBodyStr)
	log.Println("info: js run contentType > ", contentType)

	var sendBodys []string
	json.Unmarshal([]byte(sendBodyStr), &sendBodys)
	for _, b := range sendBodys {
		httpPostJson(u, contentType, b)
	}

	//log.Println("info: sendBodys > ", reflect.TypeOf(sendBodys[0]))

	requestReturn(w, 200, "ok") // 这个写入到 w 的是输出到客户端的
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LstdFlags | log.Lshortfile)
	var p string
	flag.StringVar(&tp, "t", "./js", "Template file directory is empty")
	flag.StringVar(&p, "p", "9090", "server port")
	flag.Parse()
	tpIsExists, tpErr := PathExists(tp)
	if !tpIsExists {
		if tpErr == nil {
			log.Fatal("Err: Template file directory is empty")
		} else {
			log.Fatal("Err: ", tpErr)
		}
	}
	runInit()
	log.Println("info: Template file directory is > ", tp)
	log.Println("info: server port is > ", p)
	http.HandleFunc("/send", send)         // 设置访问的路由
	err := http.ListenAndServe(":"+p, nil) // 设置监听的端口
	if err != nil {
		log.Fatal("Err: ListenAndServe ", err)
	}
}
