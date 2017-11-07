package main

import (
	"goim/libs/hash/cityhash"
	"time"

	log "github.com/thinkboy/log4go"
)

var (
	maxInt        = 1<<31 - 1
	emptyJSONBody = []byte("{}")
)

type ServerOptions struct {
	CliProto         int
	SvrProto         int
	HandshakeTimeout time.Duration
	TCPKeepalive     bool
	TCPRcvbuf        int
	TCPSndbuf        int
}

type Server struct {
	Stat      *Stat
	Buckets   []*Bucket // subkey bucket
	bucketIdx uint32
	round     *Round // accept round store
	operator  Operator
	Options   ServerOptions
}

// NewServer returns a new Server.
func NewServer(st *Stat, b []*Bucket, r *Round, o Operator, options ServerOptions) *Server {
	s := new(Server)
	s.Stat = st
	s.Buckets = b
	s.bucketIdx = uint32(len(b))
	s.round = r
	s.operator = o
	s.Options = options
	return s
}

func (server *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % server.bucketIdx
	if Debug {
		log.Debug("\"%s\" hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return server.Buckets[idx]
}

func (server *Server) RoomIds() []int32 {
	var (
		roomId  int32
		bucket  *Bucket
		roomIds = make(map[int32]struct{})
	)
	for _, bucket = range server.Buckets {
		for roomId, _ = range bucket.Rooms() {
			roomIds[roomId] = struct{}{}
		}
	}

	ids := make([]int32, 0, len(roomIds))

	for id := range roomIds {
		ids = append(ids, id)
	}

	return ids
}
