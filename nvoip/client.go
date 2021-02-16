package nvoip

import (
	"strconv"
	"strings"

	"git.resultys.com.br/lib/lower/convert/decode"
	"git.resultys.com.br/lib/lower/net/request"
	"git.resultys.com.br/sdk/nvoip-golang/queuetime"
)

type AudioParameter struct {
	TextOrAudioUrl string `json:"audio"`
	Position       int    `json:"positionAudio"`
}

type DtmfParameter struct {
	TextOrAudioUrl string `json:"audio"`
	Position       int    `json:"positionAudio"`
	MaxTime        string `json:"timedtmf"`
	Timeout        string `json:"timeout"`
	MinNumberKey   string `json:"min"`
	MaxNumberKey   string `json:"max"`
}

type RequestParameter struct {
	PhoneFrom string           `json:"caller"`
	PhoneTo   string           `json:"called"`
	Audios    []AudioParameter `json:"audios"`
	Dtmf      []DtmfParameter  `json:"dtmf"`
}

type ResponseParameter struct {
	UUID      string `json:"uuid"`
	Status    string `json:"status"`
	PhoneFrom string `json:"caller"`
	PhoneTo   string `json:"called"`
	Dtmf      string `json:"dtmf"`
}

// Client ...
type Client struct {
	AccessToken string
	CallerID    string
	qtime       *queuetime.QueueTime
}

func New(accessToken string, totalRequestBySeconds int) *Client {
	tx := float64(1) / float64(totalRequestBySeconds)
	interval := int(tx * float64(1000))

	qtime := queuetime.New(interval)
	qtime.Run()

	return &Client{
		AccessToken: accessToken,
		qtime:       qtime,
	}
}

func (c *Client) CallQueued(param RequestParameter, fn func(*ResponseParameter, error)) {
	c.qtime.Push(func() {
		fn(c.Call(param))
	})
}

func (c *Client) Call(param RequestParameter) (*ResponseParameter, error) {
	if len(param.PhoneFrom) == 0 {
		param.PhoneFrom = c.CallerID
	}

	rq := request.New("https://api.nvoip.com.br/v1/torpedo/dtmf/dynamic")
	rq.SetTimeout(5 * 60)
	rq.AddHeader("Accept", "application/json")
	rq.AddHeader("Content-Type", "application/json")
	rq.AddHeader("token_auth", c.AccessToken)

	response, err := rq.PostJSON(param)
	if err != nil {
		return nil, err
	}

	protocol := &ResponseParameter{}
	decode.JSON(response, &protocol)

	return protocol, nil
}

func (r *ResponseParameter) KeyPressed() int {
	ks := strings.ReplaceAll(r.Dtmf, "[", "")
	ks = strings.ReplaceAll(ks, "]", "")

	ki, err := strconv.Atoi(ks)
	if err != nil {
		return 0
	}

	return ki
}
