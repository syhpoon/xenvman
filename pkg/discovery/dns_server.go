/*
 MIT License

 Copyright (c) 2018 Max Kuznetsov <syhpoon@syhpoon.ca>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/syhpoon/xenvman/pkg/logger"
)

var dnsLog = logger.GetLogger("xenvman.pkg.discovery.dns_server")

type DnsServerParams struct {
	Addr string
	// Domain -> IP address
	DomainMap map[string]string
	Recursors []string
	OwnDomain string
	Ctx       context.Context
}

type DnsServer struct {
	params    DnsServerParams
	domainMap map[string]string
	client    *dns.Client

	sync.RWMutex
}

func NewDnsServer(params DnsServerParams) *DnsServer {
	return &DnsServer{
		params:    params,
		domainMap: params.DomainMap,
		client:    &dns.Client{SingleInflight: true},
	}
}

func (srv *DnsServer) Run(wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	dns.HandleFunc(".", srv.handle)

	dnsLog.Infof("Starting DNS server at %s", srv.params.Addr)

	server := dns.Server{
		Addr: srv.params.Addr,
		Net:  "udp",
	}

	go func() {
		<-srv.params.Ctx.Done()

		_ = server.Shutdown()
	}()

	if err := server.ListenAndServe(); err != nil {
		errCh <- errors.Wrapf(err, "Error running DNS server")
	}
}

func (srv *DnsServer) handle(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	dnsLog.Debugf("Got request: %+v", r)

	switch r.Opcode {
	case dns.OpcodeQuery:
		srv.processQuery(&msg)
	default:
		dnsLog.Warningf("Unable to answer request opcode=%v", r.Opcode)
	}

	_ = w.WriteMsg(&msg)
}

func (srv *DnsServer) processQuery(msg *dns.Msg) {
	for _, q := range msg.Question {
		switch q.Qtype {
		case dns.TypeA:
			if rr, err := srv.processA(q, msg); err != nil {
				dnsLog.Errorf("Error processing request: %+v", err)
			} else if rr != nil {
				msg.Answer = append(msg.Answer, rr)
			}
		default:
			if rr, err := srv.recurse(q); err != nil {
				dnsLog.Errorf("Error forwarding request: %+v", err)
			} else if rr != nil {
				msg.Answer = append(msg.Answer, rr)
			}
		}
	}
}

func (srv *DnsServer) processA(q dns.Question, msg *dns.Msg) (dns.RR, error) {
	name := q.Name

	srv.RLock()
	addr, ok := srv.domainMap[name]
	srv.RUnlock()

	if ok {
		return dns.NewRR(fmt.Sprintf("%s A %s", name, addr))
	} else if strings.HasSuffix(name, srv.params.OwnDomain) {
		dnsLog.Warningf("Internal domain %s not found", name)

		return nil, nil
	} else {
		return srv.recurse(q)
	}
}

func (srv *DnsServer) recurse(question dns.Question) (dns.RR, error) {
	m := &dns.Msg{Question: []dns.Question{question}}

	dnsLog.Debugf("Recursing request: %+v", question)

	var err error
	var resp *dns.Msg

	for _, rec := range srv.params.Recursors {
		resp, _, err = srv.client.ExchangeContext(srv.params.Ctx, m, rec)

		if err != nil {
			continue
		}

		if len(resp.Answer) > 0 {
			dnsLog.Debugf("Got response from recursor %s: %+v",
				rec, resp.Answer[0])

			return resp.Answer[0], nil
		}
	}

	return nil, errors.Wrapf(err, "Error recursing request")
}

func (srv *DnsServer) updateDomains(domains map[string]string) {
	srv.Lock()

	for dom, ip := range domains {
		srv.domainMap[dom] = ip

		dnsLog.Infof("Updating domain %s -> %s", dom, ip)
	}

	srv.Unlock()

}

func (srv *DnsServer) deleteDomains(domains []string) {
	srv.Lock()

	for _, dom := range domains {
		delete(srv.domainMap, dom)

		dnsLog.Infof("Deleting domain %s", dom)
	}

	srv.Unlock()

}

func (srv *DnsServer) getDomains() map[string]string {
	srv.RLock()
	b, _ := json.Marshal(srv.domainMap)
	srv.RUnlock()

	cp := map[string]string{}

	_ = json.Unmarshal(b, &cp)

	return cp
}
