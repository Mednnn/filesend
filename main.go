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
	f, _ := os.Create(fmt.Sprintf("pprof_%s", mode))

	switch mode {
	case "server":
		//http.ListenAndServe("0.0.0.0:8000", nil)
		srv := coalago.NewServer()
		fcontent, err := os.ReadFile("testfile20mb")
		if err != nil {
			print(err)
		}
		pprof.StartCPUProfile(f)
		srv.AddGETResource("/file", func(message *coalago.CoAPMessage) *coalago.CoAPResourceHandlerResult {
			return coalago.NewResponse(coalago.NewBytesPayload(fcontent), coalago.CoapCodeContent)
		})
		go srv.Listen(":1111")
		print("start wait")
		time.Sleep(time.Second * 20)
		print("stop")
		pprof.StopCPUProfile()

	case "client":

		cl := coalago.NewClient()
		var resp *coalago.Response
		var err error
		pprof.StartCPUProfile(f)
		resp, err = cl.GET("coap://127.0.0.1:1111/file")
		pprof.StopCPUProfile()
		if err != nil {
			print(err)
		}
		os.WriteFile("testfileDone", resp.Body, 7777)
	}

}
