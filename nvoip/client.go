package nvoip

import (
	"strconv"
	"strings"
	"time"

	"github.com/GeoinovaDev/lower-resultys/convert/decode"
	"github.com/GeoinovaDev/lower-resultys/net/request"
	"github.com/GeoinovaDev/nvoip-resultys/queuecapacity"
	"github.com/GeoinovaDev/nvoip-resultys/queuetime"
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

// AuthResponse ...
type AuthResponse struct {
	AcessToken   string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expires      int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// Client ...
type Client struct {
	NumberSip string
	UserToken string
	CallerID  string
	Timeout   int

	accessToken string
	qtime       *queuetime.QueueTime
	qcapacity   *queuecapacity.QueueCapacity
}

// New ...
func New(numberSip string, userToken string, totalRequestParallel int, totalRequestBySeconds int) *Client {
	tx := float64(1) / float64(totalRequestBySeconds)
	interval := int(tx * float64(1000))

	client := &Client{
		NumberSip: numberSip,
		UserToken: userToken,
		qtime:     queuetime.New(interval),
		qcapacity: queuecapacity.New(totalRequestParallel),
	}

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

	go client.authWorker(func() {
		client.qtime.Run()
		client.qcapacity.Run()
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

	rq := request.New("https://api.nvoip.com.br/v2/torpedo/voice")
	if c.Timeout > 0 {
		rq.SetTimeout(c.Timeout)
	}
	rq.AddHeader("Accept", "application/json")
	rq.AddHeader("Content-Type", "application/json")
	rq.AddHeader("Authorization", "Bearer "+c.accessToken)

	response, err := rq.PostJSON(param)
	if err != nil {
		return nil, err
	}

	protocol := &ResponseParameter{}
	decode.JSON(response, &protocol)

	return protocol, nil
}

func (c *Client) authWorker(fn func()) {
	isFirstLoop := true

	for {
		auth, err := c.requestNewToken()
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		c.accessToken = auth.AcessToken
		if isFirstLoop {
			fn()
		}
		isFirstLoop = false

		time.Sleep(time.Duration(auth.Expires-1000) * time.Second)
	}
}

func (c *Client) requestNewToken() (*AuthResponse, error) {
	rq := request.New("https://api.nvoip.com.br/v2/oauth/token")
	rq.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	rq.AddHeader("Authorization", "Basic TnZvaXBBcGlWMjpUblp2YVhCQmNHbFdNakl3TWpFPQ==")
	data := map[string]string{}
	data["username"] = c.NumberSip
	data["password"] = c.UserToken
	data["grant_type"] = "password"

	response, err := rq.Post(data)
	if err != nil {
		return nil, err
	}

	protocol := &AuthResponse{}
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
