package main

import (
	"fmt"

	"github.com/daviddengcn/go-colortext"
	"github.com/dixonwille/wmenu"
	wlog "gopkg.in/dixonwille/wlog.v2"
)

// newMenu creates a new menu with all the ecs containers accessible
func newMenu(sshConf SSHConf, containers []*ECSContainer) (*wmenu.Menu, error) {
	//setup menu
	menu := wmenu.NewMenu("What container do you want to `ssh` into?")
	menu.AddColor(wlog.Color{
		Code: ct.Green,
	}, wlog.Color{
		Code: ct.Yellow,
	}, wlog.Color{
		Code: ct.Magenta,
	}, wlog.Color{
		Code: ct.Yellow,
	})

	//add options
	for _, container := range containers {
		opt := fmt.Sprintf("Access container id %v on host %v running task %v", container.DockerCID[:12], container.PrivateIP, container.TaskName)
		menu.Option(opt, container, false, nil)
	}

	menu.Action(func(opts []wmenu.Opt) error {
		container, ok := opts[0].Value.(*ECSContainer)
		if !ok {
			return fmt.Errorf("error casting container back to its ECSContainer original type")
		}

		err := sshIn(sshConf, container)
		if err != nil {
			return fmt.Errorf("error reading ssh private key, got %v", err)
		}

		return nil
	})

	return menu, nil
}
