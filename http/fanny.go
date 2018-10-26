package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type MyKey struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

func main() {
	//test
	http.HandleFunc("/api/download", beginDown)
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8888", nil)

	if err != nil {
		log.Fatal("listen and server:", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("fanny.html")
	t.Execute(w, nil)
}

func beginDown(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	fp, _, err := r.FormFile("file")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer fp.Close()

	//创建 images 文件夹
	if exists, _ := checkIsExists("images"); !exists {
		os.Mkdir("images", os.ModePerm)
	}

	br := bufio.NewReader(fp)

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

	compress()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=images.zip")
	fileinfo, err := os.Stat("images.zip")
	w.Header().Set("Content-Length", strconv.FormatInt(fileinfo.Size(), 10))
	//myfp, err := os.OpenFile("images.zip", os.O_RDONLY, os.ModePerm)
	path := "images.zip"
	//io.Copy(w, myfp)
	//err :=
	http.ServeFile(w, r, path)
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
	//image_name := mykey.Key + "." + ext
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

func compress() {
	os.Remove("images.zip")
	f, err := os.OpenFile("images.zip", os.O_CREATE|os.O_WRONLY, 0666)
	zipw := zip.NewWriter(f)

	defer f.Close()
	defer zipw.Close()

	//遍历 images 文件夹
	dir, err := os.Open("images")
	if err != nil {
		log.Fatal(err)
	}

	defer dir.Close()

	files, _ := dir.Readdir(0)

	for _, file := range files {
		path := "images/" + file.Name()
		header, _ := zip.FileInfoHeader(file)
		header.Method = zip.Deflate
		writer, _ := zipw.CreateHeader(header)
		fp, _ := os.Open(path)
		io.Copy(writer, fp)
		os.Remove(path)
	}
}
