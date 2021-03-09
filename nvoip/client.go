package nvoip

import (
	"strconv"
	"strings"

	"git.resultys.com.br/lib/lower/convert/decode"
	"git.resultys.com.br/lib/lower/net/request"
	"git.resultys.com.br/sdk/nvoip-golang/queuecapacity"
	"git.resultys.com.br/sdk/nvoip-golang/queuetime"
)

// AudioParameter ...
type AudioParameter struct {
	TextOrAudioUrl string `json:"audio"`
	Position       int    `json:"positionAudio"`
}

// DtmfParameter ...
type DtmfParameter struct {
	TextOrAudioUrl string `json:"audio"`
	Position       int    `json:"positionAudio"`
	MaxTime        string `json:"timedtmf"`
	Timeout        string `json:"timeout"`
	MinNumberKey   string `json:"min"`
	MaxNumberKey   string `json:"max"`
}

// RequestParameter ...
type RequestParameter struct {
	PhoneFrom string           `json:"caller"`
	PhoneTo   string           `json:"called"`
	Audios    []AudioParameter `json:"audios"`
	Dtmf      []DtmfParameter  `json:"dtmf"`
}

// ResponseParameter ...
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
	Timeout     int
	qtime       *queuetime.QueueTime
	qcapacity   *queuecapacity.QueueCapacity
}

// New ...
func New(accessToken string, totalRequestParallel int, totalRequestBySeconds int) *Client {
	tx := float64(1) / float64(totalRequestBySeconds)
	interval := int(tx * float64(1000))

	client := &Client{
		AccessToken: accessToken,
		qtime:       queuetime.New(interval),
		qcapacity:   queuecapacity.New(totalRequestParallel),
	}

	client.qtime.Run()
	client.qcapacity.Run()

	client.qcapacity.OnPush(func(item *queuecapacity.QueueItem) {
		client.qtime.Push(func() {
			context := item.Context.(map[string]interface{})
			param := context["param"].(RequestParameter)
			fn := context["fn"].(func(*ResponseParameter, error))

			response, err := client.Call(param)
			client.qcapacity.RemoveItem(item.ID)
			fn(response, err)
		})
	})

	return client
}

// CallQueued ...
func (c *Client) CallQueued(param RequestParameter, fn func(*ResponseParameter, error)) {
	context := make(map[string]interface{})
	context["param"] = param
	context["fn"] = fn

	c.qcapacity.AddItem(context)
}

// Call ...
func (c *Client) Call(param RequestParameter) (*ResponseParameter, error) {
	if len(param.PhoneFrom) == 0 {
		param.PhoneFrom = c.CallerID
	}

	rq := request.New("https://api.nvoip.com.br/v1/torpedo/dtmf/dynamic")
	if c.Timeout > 0 {
		rq.SetTimeout(c.Timeout)
	}
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
