# goblin

> A pastebin web application focused at CLI support (sprunge inspired)



## Installation

#### Install Required Packages

```bash
$ go get github.com/kinghrothgar/goblin github.com/bmizerany/pat github.com/grooveshark/golib/gslog bitbucket.org/kardianos/osext github.com/mediocregopher/flagconfig github.com/kinghrothgar/redis/pool github.com/mediocregopher/radix/redis github.com/kinghrothgar/pygments github.com/kinghrothgar/goblin
```

#### Install Redis

Check out http://redis.io/download


#### Install Pygments (Optional)

```bash
$ pip install pygments
```
Note: depending on your configuration, pygments may install at `/usr/bin/pygmentize` or `/usr/local/bin/pygmentize`


## Configuration

```bash
$ go run goblin.go --example > goblin.conf
```



## Running

#### Start redis

```bash
$ redis-server --port 6666
```

#### Run goblin

```bash
$ go run goblin.go --conf=goblin.conf
```

or without a conf file

```bash
$ go run goblin.go --storetype=redis --storeconf=0.0.0.0:6666 --domain=0.0.0.0 --pygmentizepath=/usr/bin/pygmentize --listen=0.0.0.0:3000 --htmltemplates=/home/vagrant/code/goblin/templates/htmlTemplates.tmpl --texttemplates=/home/vagrant/code/goblin/templates/textTemplates.tmpl --staticpath=/home/vagrant/code/goblin/static
```


