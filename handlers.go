package main

import (
	"crypto/tls"
	"github.com/miekg/dns"
	"log"
	"strings"
	"time"
)

var DnsExchangeHandler *DnsHandler

type DnsHandler struct {
	clients map[string][]*dns.Client
	msgChan chan *DnsExchangeMessage
}
type DnsExchangeMessage struct {
	Message     *dns.Msg
	ReturnChan  chan *dns.Msg
	returnCount int
}

func NewDnsHandler(NameServerAddrs []string) *DnsHandler {
	res := &DnsHandler{}
	res.msgChan = make(chan *DnsExchangeMessage, len(NameServerAddrs))
	res.clients = make(map[string][]*dns.Client)

	for _, srvAddr := range NameServerAddrs {
		net := "udp"
		addr := srvAddr
		tlsServerName := ""
		if idx := strings.Index(srvAddr, ":853"); idx != -1 {
			net = "tcp-tls"
			addr = srvAddr[:idx+4]
			tlsServerName = srvAddr[idx+5:]
		}
		res.clients[addr] = make([]*dns.Client, 2)
		for i := 0; i < len(res.clients[addr]); i++ {
			res.clients[addr][i] = &dns.Client{
				Net:          net,
				ReadTimeout:  time.Second * 1,
				WriteTimeout: time.Second * 1,
				TLSConfig: &tls.Config{
					ServerName:         tlsServerName,
					InsecureSkipVerify: false,
				},
			}
			res.runWorker(res.clients[addr][i], addr)
		}
	}
	return res
}

func (h *DnsHandler) Handle(exchangeMessage *DnsExchangeMessage) {
	h.msgChan <- exchangeMessage
}

func (h *DnsHandler) runWorker(client *dns.Client, srvAddr string) {
	go func() {
		for msg := range h.msgChan {
			in, _, err := client.Exchange(msg.Message, srvAddr)
			if err != nil {
				if msg.returnCount < 3 {
					log.Printf("DNS[%s] Exchange error[%d]: %s", srvAddr, msg.returnCount, err)
					msg.returnCount++
					h.msgChan <- msg // return msg
				} else {
					log.Printf("DNS[%s] Exchange[error]: %s", srvAddr, err)
				}
				continue
			}
			if in != nil && in.Rcode != dns.RcodeSuccess {
				if in.Rcode == dns.RcodeServerFailure {
					continue
				}
			}
			addResolvedByAnswer(srvAddr, err, in.Question[0].Name, in)
			msg.ReturnChan <- in
		}
	}()
}
