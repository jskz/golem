/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	copyoverEnv                = "GOLEM_COPYOVER"
	copyoverStatePathEnv       = "GOLEM_COPYOVER_STATE"
	copyoverMinimumInheritedFD = 3
)

type copyoverState struct {
	Port        int                   `json:"port"`
	ListenerFD  int                   `json:"listenerFd"`
	InitiatedBy string                `json:"initiatedBy"`
	Clients     []copyoverClientState `json:"clients"`
}

type copyoverClientState struct {
	FD         int    `json:"fd"`
	Name       string `json:"name"`
	RemoteAddr string `json:"remoteAddr"`
}

type preparedCopyover struct {
	state          *copyoverState
	files          []*os.File
	playingClients []*Client
	loginClients   []*Client
}

type fileListener interface {
	File() (*os.File, error)
}

type fileConn interface {
	File() (*os.File, error)
}

func copyoverStateFromEnvironment() (*copyoverState, error) {
	statePath := os.Getenv(copyoverStatePathEnv)
	if statePath == "" {
		if os.Getenv(copyoverEnv) != "" {
			return nil, errors.New("copyover requested without a state file path")
		}

		return nil, nil
	}

	stateBytes, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	if err := os.Remove(statePath); err != nil {
		log.Printf("Warning: failed to remove copyover state file %s: %v.\r\n", statePath, err)
	}

	state := &copyoverState{}
	err = json.Unmarshal(stateBytes, state)
	if err != nil {
		return nil, err
	}

	if state.ListenerFD < copyoverMinimumInheritedFD {
		return nil, fmt.Errorf("invalid inherited listener fd %d", state.ListenerFD)
	}

	return state, nil
}

func openListener(state *copyoverState) (net.Listener, error) {
	if state == nil {
		return net.Listen("tcp", fmt.Sprintf(":%d", Config.Port))
	}

	listenerFile := os.NewFile(uintptr(state.ListenerFD), "copyover-listener")
	if listenerFile == nil {
		return nil, fmt.Errorf("unable to use inherited listener fd %d", state.ListenerFD)
	}
	defer listenerFile.Close()

	listener, err := net.FileListener(listenerFile)
	if err != nil {
		return nil, fmt.Errorf("recover listener from fd %d: %w", state.ListenerFD, err)
	}

	return listener, nil
}

func (game *Game) copyover(ch *Character) error {
	if ch == nil || ch.Client == nil {
		return errors.New("copyover requires a connected administrator")
	}

	prepared, err := game.prepareCopyover(ch)
	if err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			closeCopyoverFiles(prepared.files)
		}
	}()

	statePath, err := writeCopyoverState(prepared.state)
	if err != nil {
		return err
	}

	if err := game.notifyCopyoverClients(ch, prepared); err != nil {
		os.Remove(statePath)
		return err
	}

	err = startCopyoverProcess(statePath, prepared.files)
	if err != nil {
		os.Remove(statePath)
		return err
	}

	return nil
}

func (game *Game) prepareCopyover(ch *Character) (*preparedCopyover, error) {
	listener, ok := game.listener.(fileListener)
	if !ok {
		return nil, fmt.Errorf("server listener does not support copyover")
	}

	listenerFile, err := listener.File()
	if err != nil {
		return nil, fmt.Errorf("duplicate listener for copyover: %w", err)
	}

	err = inheritAcrossExec(listenerFile)
	if err != nil {
		listenerFile.Close()
		return nil, fmt.Errorf("prepare listener for copyover exec: %w", err)
	}

	prepared := &preparedCopyover{
		state: &copyoverState{
			Port:        Config.Port,
			ListenerFD:  int(listenerFile.Fd()),
			InitiatedBy: ch.Name,
		},
		files: []*os.File{listenerFile},
	}

	success := false
	defer func() {
		if !success {
			closeCopyoverFiles(prepared.files)
		}
	}()

	for client := range game.clients {
		if client.Character == nil || client.ConnectionState < ConnectionStatePlaying {
			prepared.loginClients = append(prepared.loginClients, client)
			continue
		}

		if !client.Character.Save() {
			return nil, fmt.Errorf("failed to save %s for copyover", client.Character.Name)
		}

		conn, ok := client.conn.(fileConn)
		if !ok {
			return nil, fmt.Errorf("connection for %s does not support copyover", client.Character.Name)
		}

		connFile, err := conn.File()
		if err != nil {
			return nil, fmt.Errorf("duplicate connection for %s: %w", client.Character.Name, err)
		}

		err = inheritAcrossExec(connFile)
		if err != nil {
			connFile.Close()
			return nil, fmt.Errorf("prepare connection for %s copyover exec: %w", client.Character.Name, err)
		}

		prepared.state.Clients = append(prepared.state.Clients, copyoverClientState{
			FD:         int(connFile.Fd()),
			Name:       client.Character.Name,
			RemoteAddr: remoteAddress(client.conn),
		})
		prepared.files = append(prepared.files, connFile)
		prepared.playingClients = append(prepared.playingClients, client)
	}

	success = true
	return prepared, nil
}

func (game *Game) notifyCopyoverClients(ch *Character, prepared *preparedCopyover) error {
	for _, client := range prepared.loginClients {
		if err := client.writeToConn([]byte("\r\nSorry, the game is rebooting. Come back in a few minutes.\r\n")); err != nil {
			log.Printf("Failed to notify login client during copyover: %v.\r\n", err)
		}
		client.Close()
	}

	message := "\r\nThe world begins to move...\r\n"
	for _, client := range prepared.playingClients {
		if client.Character != nil && client.Character.outputHead > 0 {
			client.displayPrompt()
			client.Character.flushOutput()
		}

		if err := client.drainSendQueue(); err != nil {
			log.Printf("Failed to drain output for %s during copyover: %v.\r\n", client.Character.Name, err)
		}

		if err := client.writeToConn([]byte(message)); err != nil {
			return fmt.Errorf("notify %s of copyover: %w", client.Character.Name, err)
		}
	}

	return nil
}

func writeCopyoverState(state *copyoverState) (string, error) {
	stateFile, err := os.CreateTemp("", "golem-copyover-*.json")
	if err != nil {
		return "", err
	}

	statePath := stateFile.Name()
	err = json.NewEncoder(stateFile).Encode(state)
	closeErr := stateFile.Close()

	if err != nil {
		os.Remove(statePath)
		return "", err
	}

	if closeErr != nil {
		os.Remove(statePath)
		return "", closeErr
	}

	return statePath, nil
}

func startCopyoverProcess(statePath string, files []*os.File) error {
	for _, file := range files {
		if file == nil || int(file.Fd()) < copyoverMinimumInheritedFD {
			return fmt.Errorf("invalid copyover file descriptor")
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	argv := make([]string, 0, len(os.Args))
	argv = append(argv, executable)
	if len(os.Args) > 1 {
		argv = append(argv, os.Args[1:]...)
	}

	return syscall.Exec(executable, argv, copyoverEnvironment(statePath))
}

func copyoverEnvironment(statePath string) []string {
	environment := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, copyoverEnv+"=") || strings.HasPrefix(entry, copyoverStatePathEnv+"=") {
			continue
		}

		environment = append(environment, entry)
	}

	environment = append(environment, copyoverEnv+"=1", copyoverStatePathEnv+"="+statePath)
	return environment
}

func closeCopyoverFiles(files []*os.File) {
	for _, file := range files {
		if file != nil {
			file.Close()
		}
	}
}

func inheritAcrossExec(file *os.File) error {
	flags, err := unix.FcntlInt(file.Fd(), unix.F_GETFD, 0)
	if err != nil {
		return err
	}

	_, err = unix.FcntlInt(file.Fd(), unix.F_SETFD, flags&^unix.FD_CLOEXEC)
	return err
}

func (game *Game) recoverCopyover(state *copyoverState) error {
	log.Printf("Copyover recovery initiated for %d client(s).\r\n", len(state.Clients))

	for _, savedClient := range state.Clients {
		if err := game.recoverCopyoverClient(savedClient); err != nil {
			log.Printf("Failed to recover copyover client %s@%s: %v.\r\n", savedClient.Name, savedClient.RemoteAddr, err)
		}
	}

	return nil
}

func (game *Game) recoverCopyoverClient(savedClient copyoverClientState) error {
	connFile := os.NewFile(uintptr(savedClient.FD), fmt.Sprintf("copyover-client-%s", savedClient.Name))
	if connFile == nil {
		return fmt.Errorf("unable to use inherited client fd %d", savedClient.FD)
	}
	defer connFile.Close()

	conn, err := net.FileConn(connFile)
	if err != nil {
		return fmt.Errorf("recover connection from fd %d: %w", savedClient.FD, err)
	}

	client := newClient(conn)
	client.ConnectionState = ConnectionStatePlaying

	ch, room, err := game.FindPlayerByName(savedClient.Name)
	if err != nil {
		client.writeToConn([]byte("\r\nYour character could not be restored after copyover. Please reconnect.\r\n"))
		client.Close()
		return err
	}

	if ch == nil {
		client.writeToConn([]byte("\r\nYour character could not be found after copyover. Please reconnect.\r\n"))
		client.Close()
		return fmt.Errorf("player not found")
	}

	client.Character = ch
	ch.Client = client
	ch.Flags |= CHAR_IS_PLAYER
	ch.Room = room
	game.clients[client] = true
	game.addPlayerCharacterToWorld(ch)

	go client.writePump(game)

	client.Send([]byte("\r\nThe uncanny disturbances settle.\r\n"))
	do_look(ch, "")

	go client.readPump(game)

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)
			if rch != ch {
				rch.Send(fmt.Sprintf("\r\n{W%s materializes.{x\r\n", ch.GetShortDescriptionUpper(rch)))
			}
		}
	}

	return nil
}
