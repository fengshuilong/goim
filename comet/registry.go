package main

import (
	"encoding/json"
	"fmt"
	inet "goim/libs/net"
	"goim/libs/net/addr"
	"goim/libs/registry"
	"goim/libs/registry/etcdv3"
	"net"
	"sort"
	"time"

	"github.com/pborman/uuid"
	log "github.com/thinkboy/log4go"
)

var (
	discovery *Discovery
	NodeID    = uuid.NewUUID().String()

	_rpcNodes       []*registry.Node
	_tcpNodes       []*registry.Node
	_websocketNodes []*registry.Node
)

// Registry Registry
type Discovery struct {
	Addrs    []string
	TTL      int
	Interval int

	Register registry.Registry
}

// InitRegistory InitRegistory
func InitRegistory(addrs []string, ttl, interval int) error {

	if len(addrs) == 0 {
		return fmt.Errorf("Registory addrs empty")
	}

	if ttl == 0 {
		ttl = 30
	}

	if interval == 0 {
		interval = 15
	}

	discovery = &Discovery{}
	r, err := etcdv3.NewRegistry(registry.Addrs(addrs...))
	if err != nil {
		return err
	}
	discovery.Register = r
	discovery.Addrs = addrs
	discovery.TTL = ttl
	discovery.Interval = interval

	return nil
}

// RegistryStartAndLoop RegistryStartAndLoop
func RegistryStartAndLoop(exit chan bool) error {

	err := registerRPC()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	err = registerTCP()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	err = registerWebsocket()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	go func() {
		t := time.NewTicker(time.Duration(discovery.Interval) * time.Second)
		for {
			select {
			case <-t.C:
				err := registerRPC()
				if err != nil {
					log.Error("RegistryStartAndLoop registerRPC err %v", err)
				}

				err = registerTCP()
				if err != nil {
					log.Error("RegistryStartAndLoop registerTCP err %v", err)
				}

				err = registerWebsocket()
				if err != nil {
					log.Error("RegistryStartAndLoop registerWebsocket err %v", err)
				}

			case <-exit:
				t.Stop()
				return

			}
		}
	}()

	return nil

}

// RegistryExit RegistryExit
func RegistryExit() {
	deregisterRPC()
	deregisterTCP()
	deregisterWebsocket()
}

func rpcServers() (*registry.Service, error) {

	name := "goim.comet.rpc"

	if _rpcNodes == nil {

		nodes := make([]*registry.Node, 0, len(Conf.RPCPushAddrs))

		for i := range Conf.RPCPushAddrs {

			_, address, err := inet.ParseNetwork(Conf.RPCPushAddrs[i])
			if err != nil {
				return nil, fmt.Errorf("registerServers inet.ParseNetwork() error(%v)", err)
			}

			ip, err := net.ResolveTCPAddr("tcp4", address)
			if err != nil {
				return nil, fmt.Errorf("inet.ParseNetwork() error(%v)", err)
			}

			address, err = addr.Extract(ip.IP.String())
			if err != nil {
				return nil, err
			}

			node := &registry.Node{
				Id:      NodeID,
				Address: address,
				Port:    ip.Port,
			}

			nodes = append(nodes, node)

		}

		_rpcNodes = nodes

	}

	// rooms
	rooms := DefaultServer.RoomIds()

	// rids := make(RoomIdSorter, 0, len(rooms))

	// for rid := range rooms {
	// 	rids = append(rids, rid)
	// }

	sort.Ints(rooms)
	sort.Sort(rids)

	roomsJSON, err := json.Marshal(rids)
	if err != nil {
		return nil, fmt.Errorf("RegistryStartAndLoop json.Marshal(rids) error(%v)", err)
	}

	s := &registry.Service{}
	s.Name = name
	s.Nodes = _rpcNodes
	s.Metadata = map[string]string{"rooms": string(roomsJSON)}

	return s, nil

}

func registerRPC() error {
	s, err := rpcServers()
	if err != nil {
		return err
	}
	return discovery.Register.Register(s, registry.RegisterTTL(time.Duration(discovery.TTL)*time.Second))
}

func deregisterRPC() error {
	s, err := rpcServers()
	if err != nil {
		return err
	}
	return discovery.Register.Deregister(s)
}

func tcpServers() (*registry.Service, error) {

	name := "goim.comet.tcp"

	if _tcpNodes == nil {

		nodes := make([]*registry.Node, 0, len(Conf.TCPBind))

		for i := range Conf.TCPBind {

			ip, err := net.ResolveTCPAddr("tcp4", Conf.TCPBind[i])
			if err != nil {
				return nil, fmt.Errorf("inet.ParseNetwork() error(%v)", err)
			}

			address, err := addr.Extract(ip.IP.String())
			if err != nil {
				return nil, err
			}

			node := &registry.Node{
				Id:      NodeID,
				Address: address,
				Port:    ip.Port,
			}

			nodes = append(nodes, node)

		}
		_tcpNodes = nodes
	}

	s := &registry.Service{}
	s.Name = name
	s.Nodes = _tcpNodes

	return s, nil
}

func registerTCP() error {
	s, err := tcpServers()
	if err != nil {
		return err
	}
	return discovery.Register.Register(s, registry.RegisterTTL(time.Duration(discovery.TTL)*time.Second))
}
func deregisterTCP() error {
	s, err := tcpServers()
	if err != nil {
		return err
	}
	return discovery.Register.Deregister(s)
}

func websocketServers() (*registry.Service, error) {

	name := "goim.comet.websocket"
	if _websocketNodes == nil {

		nodes := make([]*registry.Node, 0, len(Conf.WebsocketBind))

		for i := range Conf.WebsocketBind {

			ip, err := net.ResolveTCPAddr("tcp4", Conf.WebsocketBind[i])
			if err != nil {
				return nil, fmt.Errorf("inet.ParseNetwork() error(%v)", err)
			}

			address, err := addr.Extract(ip.IP.String())
			if err != nil {
				return nil, err
			}

			node := &registry.Node{
				Id:      NodeID,
				Address: address,
				Port:    ip.Port,
			}

			nodes = append(nodes, node)

		}

		_websocketNodes = nodes
	}

	s := &registry.Service{}
	s.Name = name
	s.Nodes = _websocketNodes

	return s, nil
}

func registerWebsocket() error {
	s, err := websocketServers()
	if err != nil {
		return err
	}
	return discovery.Register.Register(s, registry.RegisterTTL(time.Duration(discovery.TTL)*time.Second))
}

func deregisterWebsocket() error {
	s, err := websocketServers()
	if err != nil {
		return err
	}
	return discovery.Register.Deregister(s)
}
