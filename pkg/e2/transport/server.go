package transport

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/e2/node"
	"github.com/ishidawataru/sctp"
	"github.com/sirupsen/logrus"
)

// Server is the E2 server.
type Server struct {
	log      *logrus.Logger
	handler  *node.Handler
	listener *sctp.SCTPListener
	wg       sync.WaitGroup
}

// NewServer creates a new E2 server.
func NewServer(log *logrus.Logger, handler *node.Handler) *Server {
	return &Server{
		log:     log,
		handler: handler,
	}
}

// Start starts the E2 server.
func (s *Server) Start(ctx context.Context, address string) error {
	s.log.Infof("Starting E2 server on %s", address)

	addr, err := sctp.ResolveSCTPAddr("sctp", address)
	if err != nil {
		return err
	}

	s.listener, err = sctp.ListenSCTP("sctp", addr)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go s.serve(ctx)

	return nil
}

// Stop stops the E2 server.
func (s *Server) Stop() {
	s.log.Info("Stopping E2 server")
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) serve(ctx context.Context) {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				s.log.WithError(err).Error("Failed to accept new connection")
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handler.Handle(conn)
		}()
	}
}
