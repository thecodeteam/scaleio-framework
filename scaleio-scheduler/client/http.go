package client

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

//Client representation of an HTTP client
type Client struct {
	streamID   string
	url        string
	masterAddr string
	masterPath string
	client     *http.Client
}

//New generates a new HTTP client
func New(addr string, path string) *Client {
	return &Client{
		url:        "http://" + addr + path,
		masterAddr: addr,
		masterPath: path,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
			},
		},
	}
}

func parsePartialURI(str string) (string, string) {
	str = xplatform.GetInstance().Str.Trim(str, " /")
	index := strings.Index(str, "/")
	master := str[:index]
	path := str[index:]
	return master, path
}

//Send will send a HTTP payload to an HTTP server
func (c *Client) Send(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Accept", "application/json") //TODO switch to protobuf at some point
	httpReq.Header.Set("User-Agent", "scaleio/0.1")
	if c.streamID != "" {
		httpReq.Header.Set("Mesos-Stream-Id", c.streamID)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		log.Errorln("HTTP ERROR:", err)
		return nil, err
	}

	if httpResp.StatusCode == http.StatusTemporaryRedirect ||
		httpResp.StatusCode == http.StatusPermanentRedirect {
		log.Warnln("Old Master:", c.masterAddr)
		master, path := parsePartialURI(httpResp.Header.Get("Location"))
		c.masterAddr = master
		c.masterPath = path
		log.Warnln("New Master:", c.masterAddr)
		c.url = "http://" + c.masterAddr + c.masterPath
		log.Warnln("New URL:", c.url)
		msg := fmt.Sprint("StatusRedirect - New master is: ", c.masterAddr)
		log.Errorln(msg)
		return nil, errors.New(msg)
	}

	streamID := httpResp.Header.Get("Mesos-Stream-Id")
	if streamID != "" {
		if c.streamID == "" {
			log.Infoln("[MESOS-STREAM-ID] Setting to", streamID)
		} else {
			log.Infoln("[MESOS-STREAM-ID]", c.streamID, "->", streamID)
		}
		c.streamID = streamID
	}
	return httpResp, nil
}
