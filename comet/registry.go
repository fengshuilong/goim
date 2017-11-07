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
	NodeID = uuid.NewUUID().String()
)

type RegistoryOptions struct {
	Addrs    []string // 连接地址
	TTL      int      // 服务注册有效期
	Interval int      // 更新频率

}

// Registory Registory
type Registory struct {
	RPCService       *registry.Service
	TCPService       *registry.Service
	WebsocketService *registry.Service

	Options RegistoryOptions
	Server  registry.Registry

	exit bool
}

// InitRegistory InitRegistory
func InitRegistory(opt RegistoryOptions) (*Registory, error) {

	if len(opt.Addrs) == 0 {
		return nil, fmt.Errorf("Registory addrs empty")
	}

	if opt.TTL == 0 {
		opt.TTL = 30
	}

	if opt.Interval == 0 {
		opt.Interval = 15
	}

	rst := &Registory{}
	rst.Options = opt

	err := rst.initRPCService()
	if err != nil {
		return nil, err
	}

	err = rst.initTCPService()
	if err != nil {
		return nil, err
	}

	err = rst.initWebsocketService()
	if err != nil {
		return nil, err
	}

	err = rst.initRegistryServer()
	if err != nil {
		return nil, err
	}

	return rst, nil
}

func (r *Registory) initRegistryServer() error {

	etcd, err := etcdv3.NewRegistry(registry.Addrs(r.Options.Addrs...))
	if err != nil {
		return err
	}
	r.Server = etcd

	return nil
}

func (r *Registory) initRPCService() error {

	nodes := make([]*registry.Node, 0, len(Conf.RPCPushAddrs))

	for i := range Conf.RPCPushAddrs {

		_, address, err := inet.ParseNetwork(Conf.RPCPushAddrs[i])
		if err != nil {
			return fmt.Errorf("registerServers inet.ParseNetwork() error(%v)", err)
		}

		ip, err := net.ResolveTCPAddr("tcp4", address)
		if err != nil {
			return fmt.Errorf("inet.ParseNetwork() error(%v)", err)
		}

		address, err = addr.Extract(ip.IP.String())
		if err != nil {
			return err
		}

		node := &registry.Node{
			Id:      NodeID,
			Address: address,
			Port:    ip.Port,
		}

		nodes = append(nodes, node)

	}

	r.RPCService = &registry.Service{
		Name:  "goim.comet.rpc",
		Nodes: nodes,
	}

	return nil

}

func (r *Registory) RegisterRPC() error {

	s := r.RPCService

	// rooms
	rooms := RoomIdSorter(DefaultServer.RoomIds())

	sort.Sort(rooms)

	rdata, _ := json.Marshal(rooms)

	s.Metadata = map[string]string{
		"rooms": string(rdata),
	}

	return r.Server.Register(s, registry.RegisterTTL(time.Duration(r.Options.TTL)*time.Second))
}

func (r *Registory) DeregisterRPC() error {
	return r.Server.Deregister(r.RPCService)
}

func (r *Registory) initTCPService() error {

	nodes := make([]*registry.Node, 0, len(Conf.TCPBind))

	for i := range Conf.TCPBind {

		ip, err := net.ResolveTCPAddr("tcp4", Conf.TCPBind[i])
		if err != nil {
			return fmt.Errorf("inet.ParseNetwork() error(%v)", err)
		}

		address, err := addr.Extract(ip.IP.String())
		if err != nil {
			return err
		}

		node := &registry.Node{
			Id:      NodeID,
			Address: address,
			Port:    ip.Port,
		}

		nodes = append(nodes, node)

	}

	r.TCPService = &registry.Service{
		Name:  "goim.comet.tcp",
		Nodes: nodes,
	}

	return nil
}

func (r *Registory) RegisterTCP() error {
	return r.Server.Register(r.TCPService, registry.RegisterTTL(time.Duration(r.Options.TTL)*time.Second))
}

func (r *Registory) DeregisterTCP() error {
	return r.Server.Deregister(r.TCPService)
}

func (r *Registory) initWebsocketService() error {

	nodes := make([]*registry.Node, 0, len(Conf.WebsocketBind))

	for i := range Conf.WebsocketBind {

		ip, err := net.ResolveTCPAddr("tcp4", Conf.WebsocketBind[i])
		if err != nil {
			return fmt.Errorf("inet.ParseNetwork() error(%v)", err)
		}

		address, err := addr.Extract(ip.IP.String())
		if err != nil {
			return err
		}

		node := &registry.Node{
			Id:      NodeID,
			Address: address,
			Port:    ip.Port,
		}

		nodes = append(nodes, node)

	}

	r.WebsocketService = &registry.Service{
		Name:  "goim.comet.websocket",
		Nodes: nodes,
	}

	return nil
}

func (r *Registory) RegisterWebsocket() error {

	return r.Server.Register(r.WebsocketService, registry.RegisterTTL(time.Duration(r.Options.TTL)*time.Second))
}

func (r *Registory) DeregisterWebsocket() error {

	return r.Server.Deregister(r.WebsocketService)
}

func (r *Registory) Stop() {
	r.exit = true
	r.DeregisterRPC()
	r.DeregisterTCP()
	r.DeregisterWebsocket()
}

func (r *Registory) StartAndLoop() error {

	err := r.RegisterRPC()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	err = r.RegisterTCP()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	err = r.RegisterWebsocket()
	if err != nil {
		return fmt.Errorf("RegistryStartAndLoop registerRPC err %v", err)
	}

	go func() {
		t := time.NewTicker(time.Duration(r.Options.Interval) * time.Second)
		for {
			if r.exit {
				t.Stop()
				return
			}

			<-t.C

			if r.exit {
				t.Stop()
				return
			}

			err := r.RegisterRPC()
			if err != nil {
				log.Error("RegistryStartAndLoop registerRPC err %v", err)
			}

			err = r.RegisterTCP()
			if err != nil {
				log.Error("RegistryStartAndLoop registerTCP err %v", err)
			}

			err = r.RegisterWebsocket()
			if err != nil {
				log.Error("RegistryStartAndLoop registerWebsocket err %v", err)
			}

		}
	}()

	return nil

}
