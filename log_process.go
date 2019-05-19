package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/influxdata/influxdb1-client/v2"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const INFLUXDB_DSN  = "http://127.0.0.1:8086@test@testpassword@test@s"
const LOG_PATH = "./access.log"

type LogProcess struct {
	chanRead  chan []byte
	chanWrite chan *Message
	read      Reader
	write     Writer
}

type Message struct {
	TimeLocal time.Time
	BytesSent int
	Path, Method, Schema, Status string
	UpstreamTime, RequestTime float64
}

type Writer interface {
	Write(chanWrite chan *Message)
}

type Reader interface {
	Read(chanRead chan []byte)
}

type ReadFromFile struct {
	filePath string // 读取文件的路径
}
type WriteToInfluxDB struct {
	influxDbDsn string // influx db source
}

func (r *ReadFromFile) Read(chanRead chan []byte) {
	// 读取模块
	// 打开文件
	f, err := os.Open(r.filePath)
	if err != nil {
		panic(fmt.Sprintf("file open error : %s", err.Error()))
	}
	// 从文件末尾逐行读取文件内容
	f.Seek(0, 2)
	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadBytes('\n')
		if err == io.EOF {
			//fmt.Println("等待日志录入....")
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("readBytes error : %s", err.Error()))
		}
		chanRead <- line[:len(line)-1]
	}
}

func (lp *LogProcess) Process() {
	/**
	  127.0.0.1 - - [06/Jun/2016:06:12:46 +0800] http "GET /index.php HTTP/1.1" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	([\d\.]+)\s+([^\[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)
	*/

	loc, _ := time.LoadLocation("Asia/shanghai")
	r := regexp.MustCompile(`([\d\.]+)\s+([^\[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)
	for v := range lp.chanRead {
		ret := r.FindStringSubmatch(string(v))
		if len(ret) != 13 {
			log.Println("FindStringSubMatch fail:", string(v))
			continue
		}

		message := &Message{}
		t, err := time.ParseInLocation("02/Jan/2006:15:04:05 + 0000", ret[3], loc)
		if err != nil {
			log.Println("ParseInLocation fail: ", err.Error(), ret[3])
			continue
		}
		message.TimeLocal = t

		byteSent, _ := strconv.Atoi(ret[7])
		message.BytesSent = byteSent

		reqSli := strings.Split(ret[5], " ")
		if len(reqSli) != 3 {
			log.Println("strings.Split fail:", ret[5])
			continue
		}
		message.Method = reqSli[0]

		u, err := url.Parse(reqSli[1])
		if err != nil {
			log.Println("url parse fail:", reqSli[1])
			continue
		}
		message.Path = u.Path

		message.Schema = ret[4]
		message.Status = ret[6]

		upStreamTime, _ := strconv.ParseFloat(ret[11], 64)
		requestTime, _ := strconv.ParseFloat(ret[12], 64)
		message.UpstreamTime = upStreamTime
		message.RequestTime = requestTime
		//fmt.Println("------ret start------")
		//fmt.Println(message)
		//fmt.Println("------ret end-------")
		lp.chanWrite <- message
	}
}

func (w *WriteToInfluxDB) Write(chanWrite chan *Message) {

	infSli := strings.Split(w.influxDbDsn, "@")

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     infSli[0],
		Username: infSli[1],
		Password: infSli[2],
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	//fmt.Println("Write 写....")
	// 写入InfluxDB模块
	for v := range chanWrite {
		//fmt.Println("数据写入ing....")
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  infSli[3], // 库名
			Precision: infSli[4], // 精度
		})
		if err != nil {
			log.Fatal(err)
		}

		// Create a point and add to batch
		tags := map[string]string{"Path": v.Path, "Method": v.Method, "Scheme": v.Schema, "Status": v.Status}
		fields := map[string]interface{}{
			"UpstreamTime": v.UpstreamTime,
			"RequestTime": v.RequestTime,
			"BytesSent": v.BytesSent,
			"Path": v.Path,
		}

		pt, err := client.NewPoint("nginx_log", tags, fields, v.TimeLocal)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		// Close client resources
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("write success...")
		//fmt.Println(v)
	}

}
func main() {

	var path, influxDsn string
	flag.StringVar(&path, "path", LOG_PATH, "read file path")
	flag.StringVar(&influxDsn, "influxDsn", INFLUXDB_DSN, "influxDb dsn")

	flag.Parse()

	reader := &ReadFromFile{
		path,
	}
	writer := &WriteToInfluxDB{
		influxDsn,
	}
	lp := LogProcess{
		make(chan []byte),
		make(chan *Message),
		reader,
		writer,
	}

	go lp.read.Read(lp.chanRead) // 逐行读取文件内容
	go lp.Process() // 处理文本
	go lp.write.Write(lp.chanWrite) // 写入数据库

	time.Sleep(2 * 1e9 * time.Second)
}