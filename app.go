package main

import (
	"github.com/urfave/cli"
)

func newApp() *cli.App {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "servicename",
			Value: "something",
		},
		cli.StringFlag{
			Name:  "cluster",
			Value: "production",
		},
		cli.StringFlag{
			Name:  "status",
			Value: "RUNNING",
		},
		cli.StringFlag{
			Name:  "username",
			Value: "whoami",
		},
		cli.StringFlag{
			Name:  "ssh_password_key",
			Value: "something",
		},
		cli.StringFlag{
			Name:  "private_key_path",
			Value: "something_private",
		},
	}

	app.Action = func(c *cli.Context) error {
		serviceName := c.String("servicename")
		cluster := c.String("cluster")
		status := c.String("status")

		containers, err := ecsStuff(serviceName, cluster, status)
		if err != nil {
			return err
		}

		username := c.String("username")
		SSHPasswordKey := c.String("ssh_password_key")
		privateKeyPath := c.String("private_key_path")
		sshConf := NewSSHConf(username, SSHPasswordKey, privateKeyPath)

		menu, err := newMenu(sshConf, containers)
		err = menu.Run()
		if err != nil {
			return err
		}

		return nil
	}

	return app
}
