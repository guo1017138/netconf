package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"time"

	netconf "github.com/nemith/go-netconf/v2"
	ncssh "github.com/nemith/go-netconf/v2/transport/ssh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user")
	}

	// ssh-agent(1) provides a UNIX socket at $SSH_AUTH_SOCK.
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("Failed to open SSH_AUTH_SOCK: %v", err)
	}

	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: usr.Username,
		Auth: []ssh.AuthMethod{
			// Use a callback rather than PublicKeys so we only consult the
			// agent once the remote server wants it.
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	transport, err := ncssh.Dial("tcp", "192.168.122.165:830", config)
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	session, err := netconf.Open(transport)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	fmt.Println(session.ServerCapabilities())

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reply, err := session.Call(ctx, &netconf.GetConfigOp{Source: "running"})
	/* reply, err := session.Call(ctx, "<get-config><source><running><running/></source></get-config>") */
	if err != nil {
		panic(err)
	}
	fmt.Println(reply.Data)
}