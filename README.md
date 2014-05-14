# goblin

> A pastebin web application focused at CLI support (sprunge inspired)

## Installation

- Install Required Packages

```
go get github.com/bmizerany/pat 
go get github.com/grooveshark/golib/gslog
go get bitbucket.org/kardianos/osext
go get github.com/mediocregopher/flagconfig
go get github.com/kinghrothgar/redis/pool
go get github.com/mediocregopher/radix/redis
go get github.com/kinghrothgar/pygments
go get github.com/kinghrothgar/goblin
```

- Install Redis

## Configuration

```
go run goblin.go --example > goblin.conf
```

## Running

- Start redis

```
redis-server
```

- Run goblin

```
go run goblin.go
```
