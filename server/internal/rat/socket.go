package rat

import (
	"sync"

	"server/internal/net"
)

type socket struct {
	client net.Client
	abort  chan bool
}

func (socket *socket) Close() error {
	var err error
	if socket.abort != nil {
		close(socket.abort)
	}

	if socket.client != nil {
		err = socket.client.Close()
	}

	return err
}

type socketList struct {
	sockets map[net.ID]*socket
	sync.Mutex
}

func newSocketList() *socketList {
	return &socketList{
		sockets: make(map[net.ID]*socket),
	}
}

func (list *socketList) Add(client net.Client) *socket {
	list.Lock()
	defer list.Unlock()

	socket := &socket{
		client: client,
		abort:  make(chan bool),
	}

	list.sockets[client.ID()] = socket
	return socket
}

func (list *socketList) Del(id net.ID) bool {
	list.Lock()
	defer list.Unlock()

	socket, ok := list.sockets[id]
	if ok != true {
		return false
	}

	socket.Close()
	delete(list.sockets, id)
	return true
}

func (list *socketList) Get(id net.ID) *socket {
	list.Lock()
	defer list.Unlock()

	socket, ok := list.sockets[id]
	if ok != true {
		return nil
	}

	return socket
}

func (list *socketList) Close() error {
	list.Lock()
	defer list.Unlock()

	for _, socket := range list.sockets {
		socket.Close()
	}

	list.sockets = nil
	return nil
}
