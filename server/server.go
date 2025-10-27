// Package server provides a configurable cmd-stream-go server implementation.
//
// The Server accepts client connections and delegates Command handling to a
// user-provided ServerDelegate.
package server

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/ymz-ncnk/jointwork-go"
)

// WorkersCount defines the default number of workers.
const WorkersCount = 8

// LostConnCallback is invoked when the server loses its connection to a client.
type LostConnCallback = func(addr net.Addr, err error)

// New creates a new server.
func New(delegate Delegate, ops ...SetOption) (s *Server) {
	return NewWithWorkers(delegate, workersFactory{}, ops...)
}

// NewWithWorkers creates a new server with the given workers factory.
func NewWithWorkers(delegate Delegate, factory WorkersFactory, ops ...SetOption) (
	s *Server,
) {
	s = &Server{
		delegate: delegate, factory: factory, options: Options{
			WorkersCount: WorkersCount,
		},
	}
	Apply(ops, &s.options)
	return
}

// Server represents a cmd-stream server.
//
// It utilizes a configurable number of Workers to manage client connections
// using a specified ServerDelegate.
type Server struct {
	delegate Delegate
	factory  WorkersFactory
	receiver *ConnReceiver
	mu       sync.Mutex
	options  Options
}

func (s *Server) ListenAndServe(addr string) (err error) {
	listener, err := makeListener(addr, s.options)
	if err != nil {
		err = NewServerError(err)
		return
	}
	return s.Serve(listener)
}

// Serve accepts and processes incoming connections on the listener using
// Workers.
//
// Each worker handles one connection at a time.
//
// This function always returns a non-nil error:
//   - If Conf.WorkersCount == 0, it returns ErrNoWorkers.
//   - If the server was shut down, it returns ErrShutdown.
//   - If the server was closed, it returns ErrClosed.
func (s *Server) Serve(listener core.Listener) (err error) {
	if s.options.WorkersCount <= 0 {
		err = NewServerError(ErrNoWorkers)
		return
	}
	conns := make(chan net.Conn, s.options.WorkersCount)
	s.setReceiver(listener, conns)
	var (
		tasks = s.makeTasks(conns, s.delegate)
		jw    = jointwork.New(tasks)
	)
	if err = jw.Run(); err == nil {
		return NewServerError(ErrShutdown)
	}
	multiErr, ok := err.(interface{ Get(i int) error })
	if !ok {
		return NewServerError(err)
	}
	firstErr := multiErr.Get(0)
	if taskErr, ok := firstErr.(*jointwork.TaskError); ok {
		return NewServerError(taskErr.Cause())
	}
	return NewServerError(firstErr)
}

// Shutdown stops the server from receiving new connections.
//
// If server is not serving returns ErrNotServing.
func (s *Server) Shutdown() (err error) {
	if !s.serving() {
		return NewServerError(ErrNotServing)
	}
	if err = s.receiver.Shutdown(); err != nil {
		return NewServerError(err)
	}
	return
}

// Close closes the server, all existing connections will be closed.
//
// If server is not serving returns ErrNotServing.
func (s *Server) Close() (err error) {
	if !s.serving() {
		return NewServerError(ErrNotServing)
	}
	if err = s.receiver.Stop(); err != nil {
		return NewServerError(err)
	}
	return
}

func (s *Server) setReceiver(listener core.Listener, conns chan net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receiver = NewConnReceiver(listener, conns, s.options.ConnReceiver...)
}

func (s *Server) makeTasks(conns chan net.Conn, delegate Delegate) (
	tasks []jointwork.Task,
) {
	workers := s.factory.New(s.options.WorkersCount, conns, delegate,
		s.options.LostConnCallback)
	tasks = make([]jointwork.Task, 0, 1+len(workers))
	tasks = append(tasks, s.receiver)
	tasks = append(tasks, workers...)
	return
}

func (s *Server) serving() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver != nil
}

func makeListener(addr string, o Options) (
	listener core.Listener, err error,
) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	listener = l.(*net.TCPListener)
	if o.TLSConfig != nil {
		listener = listenerAdapter{tls.NewListener(l, o.TLSConfig), listener}
	}
	return
}

type listenerAdapter struct {
	net.Listener
	l core.Listener
}

func (a listenerAdapter) SetDeadline(tm time.Time) error {
	return a.l.SetDeadline(tm)
}
