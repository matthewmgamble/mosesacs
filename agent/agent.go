package main

import (
	"flag"
	"github.com/lucacervasio/mosesacs/client"
	"strconv"
)

var addr string

func init() {
	var port = flag.Int("port", 7547, "")
	flag.Parse()
	addr = ":" + strconv.Itoa(*port)
}

func main() {

	agent := cwmpclient.NewClient()

	//agent.AcsUrl = ""
	agent.Cpe.SerialNumber = "LMT3313-3"
	agent.Cpe.SoftwareVersion = "4.0.8.17785"

	agent.SetHTTPListenAddr(addr)
	agent.Run()
}
