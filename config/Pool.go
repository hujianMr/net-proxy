package config

import (
	"github.com/silenceper/pool"
	"log"
	"net"
	"time"
)

func InitPool(addr string) pool.Pool {
	poolConfig := &pool.Config{
		InitialCap: 2,
		MaxCap:     20,
		Factory: func() (interface{}, error) {
			log.Printf("dial addr = %s\n", addr)
			return net.Dial("tcp", addr)
		},
		Close: func(v interface{}) error {
			return v.(net.Conn).Close()
		},
		//Ping:        nil,
		IdleTimeout: 15 * time.Second,
	}
	var p pool.Pool
	var err error
	if p, err = pool.NewChannelPool(poolConfig); err != nil {
		log.Printf("new channel pool err = %s\n", err.Error())
		return nil
	}
	return p
}
