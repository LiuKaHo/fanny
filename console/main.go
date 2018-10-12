package console

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

type MyKey struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

func main() {
	//f, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)

	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer f.Close()

	//log.Println("CPU Profile started")
	//pprof.StartCPUProfile(f)

	//defer pprof.StopCPUProfile()

	//hf, err := os.OpenFile("heap.prof", os.O_RDWR|os.O_CREATE, 0644)

	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer hf.Close()

	//pprof.WriteHeapProfile(hf)

	//获取文件内容
	//开启多核下载
	fmt.Printf("启用 %d 核心下载\n\r", runtime.NumCPU())
	//runtime.GOMAXPROCS(runtime.NumCPU())
	if isExist, err := checkIsExists("images.txt"); !isExist {
		fmt.Println(err)
	} else {
		fp, _ := os.Open("images.txt")

		defer fp.Close()
		br := bufio.NewReader(fp)

		//新建图片目录
		if isExist, err = checkIsExists("images"); !isExist {
			err = os.Mkdir("images", os.ModePerm)
			if err != nil {
				fmt.Println("请联系程序员:")
				fmt.Println(err)
			}
		}
		var wg sync.WaitGroup

		for {
			n, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}

			if len(n) != 0 {
				wg.Add(1)
				url := string(n)
				go func() {
					//download image
					defer wg.Done()
					downloadImage(url)

				}()

			}

		}
		wg.Wait()
		fmt.Println("done, 3秒后退出")
		time.Sleep(time.Second * 3)
	}

}

func checkIsExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			myerr := fmt.Errorf(path + "is not exists")
			return false, myerr
		} else {
			myerr := fmt.Errorf("请联系程序员:%v", err)
			return false, myerr
		}
	} else {
		return true, nil
	}
}

func downloadImage(url string) {
	//开启
	url_array := strings.Split(url, `/`)
	key := len(url_array) - 2
	image_id := url_array[key]
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("请找程序员:%v", err)
	}

	body := res.Body
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("请找程序员:%v", err)
	}
	//匹配json
	regexp_string := `{\"pin\_id\":` + image_id + `,.*?\]}*`
	re, _ := regexp.Compile(regexp_string)

	one := re.FindString(string(data))
	//匹配key
	key_regexp_string := `\"key\":\".*?\", \"type\":\".*?\"`

	key_re, _ := regexp.Compile(key_regexp_string)
	key_one := "{" + key_re.FindString(one) + "}"
	var mykey MyKey
	json.Unmarshal([]byte(key_one), &mykey)

	head := "http://img.hb.aicdn.com/"
	image_url := head + mykey.Key
	ext := getExt(mykey.Type)

	image_path := "images/" + mykey.Key + "." + ext
	//获取图片数据
	fmt.Println("开始下载图片:" + url)
	image_res, _ := http.Get(image_url)
	defer image_res.Body.Close()

	image_data, _ := ioutil.ReadAll(image_res.Body)
	//写入文件
	fp, err := os.OpenFile(image_path, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer fp.Close()
	if err != nil {
		fmt.Println("请找程序员:%v", err)
	}
	fp.Write(image_data)
	fmt.Println("下载图片" + image_path + "完成")
}

func getExt(filetype string) string {
	switch filetype {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	default:
		return "jpg"
	}
}
