package tos

type Proxy struct {
	proxyHost     string
	proxyUserName string
	proxyPassword string
	proxyPort     int
}

func NewProxy(proxyHost string, proxyPort int) *Proxy {
	return &Proxy{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
	}
}

func (p *Proxy) WithAuth(username string, password string) {
	p.proxyUserName = username
	p.proxyPassword = password
}
