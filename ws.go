// package ws_for_micro provides building simple websocket server with message push.
package ws_for_micro

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	serverDefaultWSPath   = "/ws"
	serverDefaultPushPath = "/push"
)

var defaultUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

// Server defines parameters for running websocket server.
type Server struct {
	// Address for server to listen on
	Addr string

	// Path for websocket request, default "/ws".
	WSPath string

	// Path for push message, default "/push".
	PushPath string

	// Upgrader is for upgrade connection to websocket connection using
	// "github.com/gorilla/websocket".
	//
	// If Upgrader is nil, default upgrader will be used. Default upgrader is
	// set ReadBufferSize and WriteBufferSize to 1024, and CheckOrigin always
	// returns true.
	Upgrader *websocket.Upgrader

	// Check token if it's valid and return userID. If token is valid, userID
	// must be returned and ok should be true. Otherwise ok should be false.
	AuthToken func(token string) (userID string, ok bool)

	// Authorize push request. Message will be sent if it returns true,
	// otherwise the request will be discarded. Default nil and push request
	// will always be accepted.
	PushAuth func(r *http.Request) bool

	WH *websocketHandler
	PH *pushHandler
}

// ListenAndServe listens on the TCP network address and handle websocket
// request.

// Push filters connections by userID and event, then write message
func (s *Server) Push(userID, event, message string) (int, error) {
	return s.PH.push(userID, event, message)
}

// Drop find connections by userID and event, then close them. The userID can't
// be empty. The event is ignored if it's empty.
func (s *Server) Drop(userID, event string) (int, error) {
	return s.WH.closeConns(userID, event)
}

// Check parameters of Server, returns error if fail.
func (s Server) check() error {
	if !checkPath(s.WSPath) {
		return fmt.Errorf("WSPath: %s not illegal", s.WSPath)
	}
	if !checkPath(s.PushPath) {
		return fmt.Errorf("PushPath: %s not illegal", s.PushPath)
	}
	if s.WSPath == s.PushPath {
		return errors.New("WSPath is equal to PushPath")
	}

	return nil
}

// NewServer creates a new Server.
func NewWs() *Server {
	b := &binder{
		userID2EventConnMap: make(map[string]*[]eventConn),
		connID2UserIDMap:    make(map[string]string),
	}

	// websocket request handler
	wh := &websocketHandler{
		upgrader: defaultUpgrader,
		binder:   b,
	}

	// push request handler
	ph := &pushHandler{
		binder: b,
	}

	return &Server{
		WSPath:   serverDefaultWSPath,
		PushPath: serverDefaultPushPath,
		WH: wh,
		PH: ph,
	}
}

func checkPath(path string) bool {
	if path != "" && !strings.HasPrefix(path, "/") {
		return false
	}
	return true
}
