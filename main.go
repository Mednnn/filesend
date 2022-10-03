package main

import (
	"fmt"
	"github.com/coalalib/coalago"
	"os"
	"runtime/pprof"
	"time"
)

import _ "net/http/pprof"

func main() {
	mode := os.Args[1]
	//mode := "client"
	f, _ := os.Create(fmt.Sprintf("pprof_%s", mode))

	switch mode {
	case "server":
		srv := coalago.NewServer()
		fcontent, err := os.ReadFile("testfile20mb")
		if err != nil {
			print(err)
		}
		println(len(fcontent))
		println("9825")
		pprof.StartCPUProfile(f)
		srv.AddGETResource("/file", func(message *coalago.CoAPMessage) *coalago.CoAPResourceHandlerResult {
			println("connect")
			return coalago.NewResponse(coalago.NewBytesPayload(fcontent), coalago.CoapCodeContent)
		})
		srv.Listen(":9825")
		print("start wait")
		time.Sleep(time.Second * 30)
		print("stop")
		pprof.StopCPUProfile()

	case "client":
		cl := coalago.NewClient()
		var resp *coalago.Response
		var err error
		pprof.StartCPUProfile(f)
		resp, err = cl.GET("coap://147.182.133.37:9825/file")
		pprof.StopCPUProfile()
		if err != nil {
			print(err.Error())
		}
		if resp != nil {
			os.WriteFile("testfileDone", resp.Body, 7777)
		}
	}

}

//9825
//147.182.133.37
//5.129.81.97
//138.197.191.160
