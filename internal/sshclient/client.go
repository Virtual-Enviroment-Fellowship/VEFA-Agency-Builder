package sshclient

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
	host   string
	user   string
	port   int
}

func Connect(host string, port int, user, password, keyPath string) (*SSHClient, error) {
	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file '%s': %w", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (password or private key required)")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("network connection error to %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ssh handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	return &SSHClient{
		client: client,
		host:   host,
		user:   user,
		port:   port,
	}, nil
}

func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) RunCommand(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)
	output := stdoutBuf.String()
	if err != nil {
		return output, fmt.Errorf("command execution error: %w (stderr: %s)", err, strings.TrimSpace(stderrBuf.String()))
	}
	return output, nil
}

func (s *SSHClient) RunCommandStream(cmd string, onLine func(line string)) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	// Stream stdout & stderr
	readOutput := func(r io.Reader) {
		buf := make([]byte, 1024)
		var lineBuf bytes.Buffer
		for {
			n, err := r.Read(buf)
			if n > 0 {
				for i := 0; i < n; i++ {
					b := buf[i]
					if b == '\n' {
						if onLine != nil {
							onLine(lineBuf.String())
						}
						lineBuf.Reset()
					} else {
						lineBuf.WriteByte(b)
					}
				}
			}
			if err != nil {
				if lineBuf.Len() > 0 && onLine != nil {
					onLine(lineBuf.String())
				}
				break
			}
		}
	}

	go readOutput(stdout)
	go readOutput(stderr)

	return session.Wait()
}

func (s *SSHClient) WriteRemoteFile(remotePath string, content string, permissions string) error {
	b64 := base64.StdEncoding.EncodeToString([]byte(content))
	dir := remotePath[0:strings.LastIndex(remotePath, "/")]
	
	cmd := fmt.Sprintf("mkdir -p %s && echo '%s' | base64 -d > %s && chmod %s %s",
		dir, b64, remotePath, permissions, remotePath)

	_, err := s.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed writing remote file %s: %w", remotePath, err)
	}
	return nil
}
