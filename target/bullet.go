package target

import "net/http"

type Bullet struct {
	Shot *Shot
	Request *http.Request
	Client *http.Client
	Transport *http.Transport
}

