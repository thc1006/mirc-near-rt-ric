// pkg/o1/netconf/server.go
package netconf

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"sync/atomic"
)

// Server is a NETCONF server.
type Server struct {
	sshConfig        *ssh.ServerConfig
	sessionIDCounter uint32
	// o1Server  *o1.O1Server // This will be used later
}

// NewServer creates a new NETCONF server.
func NewServer() (*Server, error) {
	config := &ssh.ServerConfig{
		// In a real implementation, you would use a more secure key exchange algorithm
		// and a proper host key.
		NoClientAuth: true, // This is for testing only.
	}

	// In a real implementation, you would load a private key from a file.
	privateBytes := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAxbXKNz3uSBKBs2VM7bNK3JqILaB5EoIQuBOW0Pd67ZVxQoJZ
wnUR1+jmZOat9BksiibCgbIx9IhhipvBh7EqQX/YGMiFfey3kYl4d7O9/mVAMQoG
hT3e2rikXnGhQaZNBgoNL7rTfST5h7UgY66zjdH0hkSwdvzATsZ5PVfyCBy76QsJ
5GJRsU+grKqDr9Gj0lY5a8OcSgfIj2FBhF8aaCkvh9OwmRkVyn96NwHVLTmBx85I
f5/XirIneD9JVVmFKJDvUD/cf6hRUTdluoqMQer5kWWm1gp3eSRDN2YBlNmBfWle
jJ4u5rYMalPeuHlUVAdTxADtFkNbjagHeRDdUwIDAQABAoIBAQC6yIsZc3XZOzqz
nCF4c6lnDstWp8OaO6zF6yPRmezV5hiRaAqazvUjkNGRQ+nVsa7FeebKlunhBrN4
Orw0kKjGJpymlVKga/HlGgXouLPnUgq6CamtWY1f/46x9xIMrqsX6IkarZs+IJ9p
mTHXWuYhNtiXvO8mCpU4FwPVc2+iPrBbQ3hxIvkTiprCDk/JRXGEINVzqp6wGyM8
9iG2Mj/CabwuX5pxgNGO65MXaf+plD5vN5sf+fWwSwG/0fipw5xHn8PenDJjeUmK
jgl8H+aSDkHjNDtCbnz7gJ08/GZj9TFlFKKlSXLltH0dsKVhZx4cMMTIIJEDmS/n
5be5uvKBAoGBAPaXT0Nci1fqAWmCTWj6bDWsNSYQITBZDwJGRVNDWobfI/JILj3o
dIqqW4fN1QDSBKEWVON2IidNB2MZYlQ6UAHgcJl1iBQGlmNIrkEt/wEHMO374WlV
bhSQG/Ej2KShAgBM7ILc9LoOSgq5zgx6nlmkfNWiNaqLtZD2S2inJHwTAoGBAM1B
A3Om9K7Bo19fam5eIeUbJm6J1uXWaUMFVOEXhJ6yOY+6KIoZqb6h8dWUYeK8ABLJ
kjF1WkFkhJl3KfXJ94oCB9GVb8PPcwy78S8f+XqCRBt2qbYL/iEFDQf2qRBklZzT
GeuGy050PZqFDjy1wFqRceuYMZ9D7Z+V/Is9h8HBAoGBAKTjstXZWTf1OvKBdh/f
bGJLf9Ku8HJy6u1bbdnqbOtI5LGLAJjXCq76kW/y/B6rUPsigqsDAH2FLY5fl/e4
cm1+1exXwiGZ8g/7tsHQ7vaSB27rgeQ8gvpvDrAlhyU2oK7wwSoUc/TBv1MCwbxo
deB9dBgqenZLK6L+fphBQ81PAoGAFekeyTXFYPJi0keJQQbjb2WakKo+OoLM8c6b
5PtxuM8lveYNddCPgj4fZsFUQbP3/gluhcEVRW3Jiehink5VEnJtCz58k9aNXYqi
kHlFVIKbaqMcMsbM9hFn9rWqDonuPrN6TN4yzcky2k/h2TE9u21TT+cLRQknUKXe
M6750wECgYEArK1b7dqbrefBWDop1eCfpwxwZdmM1bFF+LZIxqEB2U07mOSm7k6c
cORpM98kp9ZBEERKoe5sSW5ioAKxubscveV755EUKZ9f3ujwpOqOwveQqOxScijV
yFVPILaOW5cWjdzjPBeOhLsM1yo75nQ07/2SHfyzyC3Ii/NXo6pF4Ik=
-----END RSA PRIVATE KEY-----`)
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	config.AddHostKey(private)

	return &Server{
		sshConfig: config,
	},
	nil
}

// Start starts the NETCONF server.
func (s *Server) Start(listener net.Listener) {
	log.Printf("NETCONF server listening on %s", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection handles a new SSH connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		log.Printf("Failed to establish SSH connection: %v", err)
		return
	}
	log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		// Handle requests on the channel
		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "netconf" {
						req.Reply(true, nil)
						// Start NETCONF session
						sessionID := atomic.AddUint32(&s.sessionIDCounter, 1)
						s.startNetconfSession(channel, sessionID)
					} else {
						req.Reply(false, nil)
					}
				default:
					req.Reply(false, nil)
				}
			}
		}(requests)
	}
}

// startNetconfSession starts a new NETCONF session.
func (s *Server) startNetconfSession(channel ssh.Channel, sessionID uint32) {
	defer channel.Close()
	log.Printf("Starting NETCONF session with ID %d", sessionID)

	// 1. Send NETCONF Hello message
	helloMsg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities>
    <capability>urn:ietf:params:netconf:base:1.0</capability>
  </capabilities>
  <session-id>%d</session-id>
</hello>
]]>]]>`, sessionID)
	_, err := channel.Write([]byte(helloMsg))
	if err != nil {
		log.Printf("Failed to send hello message: %v", err)
		return
	}
	log.Println("Sent NETCONF Hello message to client.")

	// 2. Read client's hello and subsequent RPCs
	// For now, just read and log. A real implementation would parse XML.
	buf := make([]byte, 4096)
	for {
		n, err := channel.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("NETCONF session closed by client.")
			} else {
				log.Printf("Error reading from NETCONF session: %v", err)
			}
			break
		}
		log.Printf("Received NETCONF message:\n%s", string(buf[:n]))
	}
}
