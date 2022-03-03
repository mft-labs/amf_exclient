/****************************************************************************
 *
 * Copyright (C) Agile Data, Inc - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by MFTLABS <code@mftlabs.io>
 *
 ****************************************************************************/
package exclient

import (
	"bufio"
	"errors"
	//"fmt"
	"os"
	"os/exec"
	"strings"
)

type ExClient struct {
	cmdConnected bool
	cmdIn        *bufio.Reader
	cmdOut       *bufio.Writer
	Pid          string
	cmd          *exec.Cmd
	controller   string
	env          string
	name         string
}

func (ex *ExClient) Init(name, controller, env string) {
	ex.controller = controller
	ex.env = env
	ex.name = name
}

func (ex *ExClient) IsConnected() bool {
	return ex.cmdConnected
}

func (ex *ExClient) Disconnect() {
	ex.cmdConnected = false
	if !ex.cmd.ProcessState.Exited() {
		cmd := exec.Command("kill", "-9", ex.Pid)
		cmd.Run()
	}
}

func (ex *ExClient) Run(cmd string) error {
	_, err := ex.cmdOut.WriteString(cmd)
	if err != nil {
		ex.cmdConnected = false
		return errors.New("Error communicating with controller [" + err.Error() + "]")
	}
	ex.cmdOut.Flush()
	return nil
}

func (ex *ExClient) GetNext() (string, error) {
	var err error
	var line string
	for {
		line, err = ex.cmdIn.ReadString('\n')
		if err != nil {
			ex.cmdConnected = false
			return "", errors.New("Error communicating with controller [" + err.Error() + "]")
		}
		if !strings.HasPrefix(line, "{\"") {
			continue
		}
		break
	}
	return line, nil
}

func (ex *ExClient) Connect() error {
	if ex.cmdConnected {
		ex.Disconnect()
	}
	ex.cmd = exec.Command("sh", ex.controller, "("+ex.name+")", "2>&1")
	ex.cmd.Env = append(os.Environ(), ex.env)
	stdin, _ := ex.cmd.StdinPipe()
	ex.cmdOut = bufio.NewWriter(stdin)
	stdout, _ := ex.cmd.StdoutPipe()
	ex.cmdIn = bufio.NewReader(stdout)
	err := ex.cmd.Start()
	if err != nil {
		return errors.New("Could not start coprocess [" + err.Error() + "]")
	}
	//	fmt.Printf("pid: %d\n", ex.cmd.ProcessState.Pid())
	//ex.cmdOut.WriteString("\n")
	//ex.cmdOut.Flush()
	cmdPid, err := ex.cmdIn.ReadString('\n')
	if err != nil {
		return errors.New("Could not read coprocess pid [" + err.Error() + "]")
	}
	ex.Pid = strings.TrimSpace(cmdPid)
	ex.cmdConnected = true
	return nil
}
