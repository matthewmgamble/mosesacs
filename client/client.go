package cwmpclient

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/lucacervasio/mosesacs/cwmp"
	"github.com/lucacervasio/mosesacs/daemon"
	"github.com/lucacervasio/mosesacs/xmpp"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

var (
	defaultHTTPListenAddr string = ":7547"
	interval              string = "60"
)

type Agent struct {
	Status         string
	AcsUrl         string
	Cpe            daemon.CPE
	HTTPListenAddr string
	T              string
}

func NewClient() (agent Agent) {
	serial := strconv.Itoa(random(1000, 5000))
	connection_request_url := "/ConnectionRequest-" + serial
	cpe := daemon.CPE{serial, "MOONAR LABS", "001309", connection_request_url, "asd", "asd", "0 BOOTSTRAP", nil, &daemon.Request{}, "4324asd", time.Now().UTC(), "TR181", false}
	//agent = Agent{"initializing", "http://localhost:9292/acs", cpe, defaultHTTPListenAddr}
	//agent = Agent{"initializing", "http://192.168.0.105:9292/acs", cpe, defaultHTTPListenAddr}

	//agent = Agent{"initializing", "http://192.168.99.100:7547", cpe, defaultHTTPListenAddr}
	//agent = Agent{"initializing", "http://211.75.3.36:7547", cpe, defaultHTTPListenAddr}
	agent = Agent{"initializing", "http://localhost:7547", cpe, defaultHTTPListenAddr, ""}
	//log.Println(agent)
	return
}

func (a Agent) SetHTTPListenAddr(addr string) {
	a.HTTPListenAddr = addr
}

func (a Agent) String() string {
	return fmt.Sprintf("Agent running with serial %s and connection request url %s\n", a.Cpe.SerialNumber, a.Cpe.ConnectionRequestURL)
}

func (a Agent) Run() {
	http.HandleFunc(a.Cpe.ConnectionRequestURL, a.connectionRequestHandler)
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		log.Println("ReceivedSetParameterValuesRequest")
		log.Println("Send SetParameterValuesRequest")
		log.Println("0")
	})
	http.HandleFunc("/periodicuploadinterval", func(w http.ResponseWriter, r *http.Request) {
		log.Println("PeriodicUploadInterval is now ", interval)
	})
	log.Println("Start http server waiting connection request")
	a.startConnection()
	//  a.startXmppConnection()

	http.ListenAndServe(a.HTTPListenAddr, nil)
}

func (a Agent) connectionRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got connection request, send Inform to %s", a.AcsUrl)
	a.startConnection()
}

func random(min, max int) int {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Intn(max-min) + min
}

func (a Agent) startXmppConnection() {
	log.Println("starting StartXmppConnection")
	xmpp.StartClient("cpe2@mosesacs.org", "password1234", func(str string) {
		log.Println("got " + str)
	})
}

func (a Agent) startConnection() {
	log.Printf("send Inform to %s", a.AcsUrl)
	var msgToSend []byte
	msgToSend = []byte(cwmp.Inform(a.Cpe.SerialNumber))

	tr := &http.Transport{}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: tr, Jar: jar}
	envelope := cwmp.SoapEnvelope{}
	u, _ := url.Parse(a.AcsUrl)

	resp, err := client.Post(a.AcsUrl, "text/xml", bytes.NewBuffer(msgToSend))
	if err != nil {
		log.Fatal("server unavailable", err)
		return
	}
	for {
		if resp.ContentLength == 0 {
			log.Println("got empty post, close connection")
			resp.Body.Close()
			tr.CloseIdleConnections()
			break
		} else {
			tmp, _ := ioutil.ReadAll(resp.Body)
			body := string(tmp)
			xml.Unmarshal(tmp, &envelope)

			if envelope.KindOf() == "GetParameterValues" {
				log.Println("Received GetParameterValuesRequest")
				log.Println("Send GetParameterValuesResponse")
				var leaves cwmp.GetParameterValues_
				xml.Unmarshal([]byte(body), &leaves)
				msgToSend = []byte(cwmp.BuildGetParameterValuesResponse(a.Cpe.SerialNumber, leaves))
			} else if envelope.KindOf() == "GetParameterNames" {
				if a.T == "get" {
					log.Println("Received GetParameterValuesRequest")
					log.Println("Send GetParameterValuesResponse")
				} else if a.T == "set" {
					log.Println("Received SetParameterValusRequest")
					log.Println("Send SetParameterValuesResponse")
				} else {
					log.Println("Received GetParameterNamesRequest")
					log.Println("Send GetParameterNamesResponse")
				}
				var leaves cwmp.GetParameterNames_
				xml.Unmarshal([]byte(body), &leaves)
				msgToSend = []byte(cwmp.BuildGetParameterNamesResponse(a.Cpe.SerialNumber, leaves))
			} else if envelope.KindOf() == "SetParameterValues" {
				//log.Println("Send SetParameterValuesResponse")
				log.Println("Received SetParameterValuesRequest")
				log.Println("Send SetParameterValuesResponse")
				var leaves cwmp.SetParameterValues_
				xml.Unmarshal([]byte(body), &leaves)
				fmt.Println(string(body))
				fmt.Println(leaves)
			} else {
				//log.Println("send empty post")
				msgToSend = []byte("")
			}

			client.Jar.SetCookies(u, resp.Cookies())
			resp, err = client.Post(a.AcsUrl, "text/xml", bytes.NewBuffer(msgToSend))
			if err != nil {
				break
			}
			//log.Println(resp.Header)
		}
	}

}
