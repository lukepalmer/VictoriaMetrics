package faststats

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/ingestserver"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/netutil"
	"github.com/VictoriaMetrics/metrics"
)

// TODO: most of 'ingestserver' is copy-pasted and faststats follows that pattern. Factor out common server code.

var (
	writeRequestsTCP = metrics.NewCounter(`vm_ingestserver_requests_total{type="FastStats", name="write"}`)
	writeErrorsTCP   = metrics.NewCounter(`vm_ingestserver_request_errors_total{type="FastStats", name="write"}`)
)

// Server accepts the FastStats protocol over TCP.
type Server struct {
	addr  string
	lnTCP net.Listener
	wg    sync.WaitGroup
	cm    ingestserver.ConnsMap
}

// MustStart starts a FastStats server on the given addr.
//
// The incoming connections are processed with insertHandler.
//
// If useProxyProtocol is set to true, then the incoming connections are accepted via proxy protocol.
// See https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
//
// MustStop must be called on the returned server when it is no longer needed.
func MustStart(addr string, useProxyProtocol bool, insertHandler func(r io.Reader) error) *Server {
	logger.Infof("starting TCP FastStats server at %q", addr)
	lnTCP, err := netutil.NewTCPListener("FastStats", addr, useProxyProtocol, nil)
	if err != nil {
		logger.Fatalf("cannot start TCP FastStats server at %q: %s", addr, err)
	}

	s := &Server{
		addr:  addr,
		lnTCP: lnTCP,
	}
	s.cm.Init("FastStats")
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.serveTCP(insertHandler)
		logger.Infof("stopped TCP FastStats server at %q", addr)
	}()
	return s
}

// MustStop stops the server.
func (s *Server) MustStop() {
	logger.Infof("stopping TCP FastStats server at %q...", s.addr)
	if err := s.lnTCP.Close(); err != nil {
		logger.Errorf("cannot close TCP FastStats server: %s", err)
	}
	s.cm.CloseAll(0)
	s.wg.Wait()
	logger.Infof("TCP FastStats servers at %q have been stopped", s.addr)
}

func (s *Server) serveTCP(insertHandler func(r io.Reader) error) {
	var wg sync.WaitGroup
	for {
		c, err := s.lnTCP.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) {
				if ne.Temporary() {
					logger.Errorf("FastStats: temporary error when listening for TCP addr %q: %s", s.lnTCP.Addr(), err)
					time.Sleep(time.Second)
					continue
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				logger.Fatalf("unrecoverable error when accepting TCP FastStats connections: %s", err)
			}
			logger.Fatalf("unexpected error when accepting TCP FastStats connections: %s", err)
		}
		if !s.cm.Add(c) {
			_ = c.Close()
			break
		}
		wg.Add(1)
		go func() {
			defer func() {
				s.cm.Delete(c)
				_ = c.Close()
				wg.Done()
			}()
			writeRequestsTCP.Inc()
			if err := insertHandler(c); err != nil {
				writeErrorsTCP.Inc()
				logger.Errorf("error in TCP FastStats conn %q<->%q: %s", c.LocalAddr(), c.RemoteAddr(), err)
			}
		}()
	}
	wg.Wait()
}
