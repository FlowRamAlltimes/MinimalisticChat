 # 📞Minimalistic Chat v1.6

Cli TCP chat based on golang
So you can download it and use now in releases. Made on x86/amd64🖥️

Become mymentor please, I really like what I do!

## Requirements

1. Golang 1.19 at least

2. Friend or somebody else and a bit of time

3. The disire to support me

## Description

The minimum-interface tcp chat. Just pet-project, nothing more.
I used RWMutex and gorutines. I hope without race conditions

## Useful commands

Client part

```bash

/health - sends you informatin about server

/list - sends you all active members

/help - shows all info

/exit - leave the server

/quit - leave the sever

/msg Nickname Message - send private message

```

Server part 

```bash

/status - all necessary information

// also you'll get logs with 

```


## How to use

Firstly you need to download files(host and client parts from releases) and after you need to make file work with:


### Linux

```bash

// Make your certificate on vps!


openssl genrsa -out ca.key 2048

openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=MyLocalCA"


sudo nano server.conf // i use nano

// Enter your config with your IP address in config, ask LLM about it


openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr -config server.conf


openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extfile server.conf -extensions v3_ext


// after you should 

// on vps

go build -o server.go

./server


// on local machine

go build -o client.go

// copy ca.crt from vps

./client

// thats all

```

I also want you to send your feedback about using this console app :3


## Star it please it will provide me for future updates 
