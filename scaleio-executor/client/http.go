package client

import (
	"bytes"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

//Client representation of an HTTP client
type Client struct {
	streamID string
	url      string
	client   *http.Client
}

//New generates a new HTTP client
func New(addr string, path string) *Client {
	return &Client{
		url: "http://" + addr + path,
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
