# gobin

> A pastebin web application focused at CLI support (sprunge inspired)



## Installation

#### Install Required Packages

```bash
$ go get github.com/kinghrothgar/gobin github.com/bmizerany/pat github.com/grooveshark/golib/gslog bitbucket.org/kardianos/osext github.com/mediocregopher/flagconfig github.com/kinghrothgar/redis/pool github.com/mediocregopher/radix/redis github.com/kinghrothgar/pygments github.com/kinghrothgar/gobin
```

#### Install Redis

Check out https://redis.io/download


#### Install Pygments (Optional)

```bash
$ pip install pygments
```
Note: depending on your configuration, pygments may install at `/usr/bin/pygmentize` or `/usr/local/bin/pygmentize`


## Configuration

```bash
$ go run gobin.go --example > gobin.conf
```



## Running

#### Start redis

```bash
$ redis-server --port 6666
```

#### Run gobin

```bash
$ go run gobin.go --conf=gobin.conf
```

or without a conf file

```bash
$ go run gobin.go --storetype=redis --storeconf=0.0.0.0:6666 --domain=0.0.0.0 --pygmentizepath=/usr/bin/pygmentize --listen=0.0.0.0:3000 --htmltemplates=/home/vagrant/code/gobin/templates/htmlTemplates.tmpl --texttemplates=/home/vagrant/code/gobin/templates/textTemplates.tmpl --staticpath=/home/vagrant/code/gobin/static
```


