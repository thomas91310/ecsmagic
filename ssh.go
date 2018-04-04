package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ScaleFT/sshkeys"
	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
)

// SSHConf defines all the necessary info to authenticate using ssh
type SSHConf struct {
	Username       string
	PasswordKey    string
	PrivateKeyPath string
}

func publicKeyFile(passwordKey, file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := sshkeys.ParseEncryptedPrivateKey(buffer, []byte(passwordKey))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func sshIn(sshConf SSHConf, container *ECSContainer) error {
	server := fmt.Sprintf("%v:22", container.PrivateIP)
	publicKey, err := publicKeyFile(sshConf.PasswordKey, sshConf.PrivateKeyPath)
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: sshConf.Username,
		Auth: []ssh.AuthMethod{
			publicKey,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", server, config)
	if err != nil {
		return fmt.Errorf("failed to dial %v, got %v", container.PrivateIP, err.Error())
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err)
	}

	session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
	in, _ := session.StdinPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO:  0,
		ssh.IGNCR: 1,
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %v", err)
	}

	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %v", err)
	}

	// Handle control + C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			<-c
			os.Exit(0)
		}
	}()

	time.Sleep(1 * time.Second)
	// cmd := fmt.Sprintf("sudo docker exec -it %v bash", container.DockerCID)

	// fmt.Fprintf(in, "")
	// fmt.Fprintf(in, cmd)
	// fmt.Fprintf(in, "\n")

	reader := bufio.NewReader(os.Stdin)
	for {
		str, _ := reader.ReadString('\n')
		fmt.Fprintf(in, str)
	}
}
