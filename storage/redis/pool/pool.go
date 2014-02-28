package pool

import (
	"github.com/mediocregopher/radix/redis"
	"github.com/grooveshark/golib/gslog"
)

// TODO: figure out channel direction
type Pool struct {
	outPool chan *Client
	inPool chan *Client
	versionChan chan string
	network, addr string
}

type Client struct {
	*redis.Client
	versionStr string
}

func buildVersionStr(network string, addr string) string {
	return network + addr
}

func New(network string, addr string, capacity int) *Pool {
	var outPool, inPool chan *Client
	if capacity < 1 {
		outPool = make(chan *Client)
		inPool = make(chan *Client)
	} else {
		outPool = make(chan *Client, capacity)
		inPool = make(chan *Client, capacity)
	}

	pool := &Pool{
		outPool: outPool,
		inPool: inPool,
		versionChan: make(chan string),
		network: network,
		addr: addr,
	}
	go poolMan(pool, buildVersionStr(network, addr))
	return pool
}

func flushChan(c chan *Client) {
	for {
		select {
		case client := <-c:
			client.Close()
			continue
		default:
			return
		}
	}
}

func poolMan(pool *Pool, versionStr string) {
	for {
		// TODO: Do I need to break? I think I do
		select {
		case str := <-pool.versionChan:
			if str != versionStr {
				gslog.Debug("POOL: flushing pool because versionStr changed")
				versionStr = str
				flushChan(pool.outPool)
			}
			continue
		case client := <-pool.inPool:
			if client.versionStr == versionStr {
				gslog.Debug("POOL: poolMan added client back to pool")
				pool.outPool <- client
			} else {
				client.Close()
				gslog.Debug("POOL: poolMan discarded client")
			}
			continue
		}
	}
}


// Get a *Client from the pool
// Returns nil on failure
func (p *Pool) Get() (*Client, error) {
	select {
	case r := <-p.outPool:
		gslog.Debug("POOL: got client from outPool")
		return r, nil
	default:
	}

	versionStr := buildVersionStr(p.network, p.addr)
	r, err := redis.Dial(p.network, p.addr)
	if err == nil {
		// TODO: learn how to use a mix of named and unamed fields
		gslog.Debug("POOL: created new client")
		return &Client{r, versionStr}, nil
	}
	return nil, err
}

// Put a *Client into the pool
// Please don't give back clients that return errors :)
// Returns false on failure
func (p *Pool) Put(client *Client) bool {
	select {
	case p.inPool <- client:
		return true
	default:
		return false
	}
}

// Configure the network and addr pool is connecting too
func (p *Pool) SetConnection(network string, addr string) {
	p.network = network
	p.addr = addr
	p.versionChan <- buildVersionStr(network, addr)
}
