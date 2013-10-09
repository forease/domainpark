domainpark
==========

A simple domain park daemon by Golang

## How to build

    git clone https://github.com/forease/domainpark.git
    cd domainpark
    go build -o domainpark *.go
    
## How to run

    ./domainpark -c serverd.conf
    
    
## How to config
    
    
    [common]
    webport = 9090 # web server port

    
    [redis]
    host = localhost # redis host
    port = 6379      # redis port
    auth = 
    db   =
    
    [smtp]
    # send report smtp setup
    daemon = false
    host = smtp.gmail.com
    port = 587
    user = test@gmail.com
    password = testmail 
    auth = true
    tls  = true

## About

    北京实易时代科技有限公司
    Beijing ForEase Times Technology Co., Ltd.
    http://www.forease.net

    
## Contact me

    Jonsen Yang
    im16hot#gmail.com (replace # with @)

