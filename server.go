package main

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"strings"
	"time"
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		//log.Printf("Query for %s as %d\n", q.Name, q.Qtype)
		processed := false
		switch q.Qtype {
		case dns.TypeA:
			for name, ip := range config.Address {
				// check q.Name has suffix name without last char in q.Name
				if strings.HasSuffix(q.Name[:len(q.Name)-1], name) {
					rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
					if err == nil {
						rr.Header().Ttl = 30
						m.Answer = append(m.Answer, rr)
						addResolvedByAnswer("config", err, name, m)
						processed = true
						continue
					}
				}
			}
		}
		if processed == false {
			cachedReq := cache.Get(q.Qtype, q.Name)
			if cachedReq != nil {
				m.Answer = cachedReq
				continue
			}

			r, err := Lookup(m)
			if err == nil {
				m.Answer = r.Answer
				cache.Set(q.Qtype, q.Name, m.Answer)
				go ipSet.Set(q.Name[:len(q.Name)-1], m.Answer)
				if err != nil {
					fmt.Printf("failed to ipSet : %v\n", err)
				}
			} else {
				fmt.Printf("failed to exchange: %v\n", err)
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	parseQuery(m)
	w.WriteMsg(m)
}

func Lookup(m *dns.Msg) (*dns.Msg, error) {

	req := new(dns.Msg)
	req.SetReply(m)
	req.Response = false

	qName := req.Question[0].Name

	c := &dns.Client{
		ReadTimeout:  time.Millisecond * 250,
		WriteTimeout: time.Millisecond * 100,
	}

	res := make(chan *dns.Msg, 1)
	queue := make(chan int, 1)
	queue <- 1
	stopped := false
	L := func(nameserver string, i int) {
		<-queue
		r, _, err := c.Exchange(req, nameserver)
		defer func() {
			if !stopped {
				queue <- 1
			}
		}()
		if err != nil {
			//log.Printf("[%d] %s socket error on %s, error: %s", i, qName, nameserver, err.Error())
			return
		}
		if r != nil && r.Rcode != dns.RcodeSuccess {
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}
		addResolvedByAnswer(nameserver, err, qName, r)
		stopped = true
		res <- r
	}

	// Start lookup on each nameserver top-down, in every second
	for i, nameserver := range config.Nameservers {
		go L(nameserver, i)
	}

	ticker := time.NewTicker(2000 * time.Millisecond)
	defer ticker.Stop()

	select {
	case r := <-res:
		return r, nil
	case <-ticker.C:
		return nil, errors.New("can't resolve ip for " + qName + " by timeout")
	}
}

func addResolvedByAnswer(nameserver string, err error, qName string, r *dns.Msg) {
	rr, err := dns.NewRR(fmt.Sprintf("%s TXT %s", "dns.resolved.via", nameserver))
	if err == nil {
		rr.Header().Ttl = 60
		r.Answer = append(r.Answer, rr)
	}
}
