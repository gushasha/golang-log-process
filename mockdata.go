package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	//127.0.0.1 - - [06/Jun/2016:06:12:46 +0800] http "GET /index.php HTTP/1.1" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	//s := (
	//	127.0.0.1 - - [06/Jun/2016:06:12:46 +0800] http "GET /index.php HTTP/1.1" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	//	)
	//
	for i := 0; i< 10 ; i++ {
		str := generateContent()
		fmt.Println(str)
		tracefile(str)
		time.Sleep(1 * time.Second)
	}

}

func generateContent() string {
	actionList := [...]string{
		"/index.php",
		"/user/detail.php",
		"/user/list.php",
		"/user/search.php",
		"/login.php",
		"/logout.php",
	}

	host := "127.0.0.1"
	data := time.Now().Format("02/Jan/2006:15:04:05 + 0000")
	method := "GET"
	action := actionList[rand.Intn(2)]
	schema := "HTTP/1.1"
	status := 200
	size := rand.Intn(10000)
	request := rand.Float64() * 2
	up := rand.Float64() * 2
	ret := fmt.Sprintf("%s - - [%s] http \"%s %s %s\" %d %d \"-\" \"KeepAliveClient\" \"-\" %f %f \n", host, data, method, action, schema, status, size, request, up)
	return ret
}

func tracefile(str_content string) {
	fd, _ := os.OpenFile("./access.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	buf := []byte(str_content)
	_, err := fd.Write(buf)
	if err != nil {
		fmt.Println("fd.Write fail: ", err.Error())
	}
	fd.Close()
}
