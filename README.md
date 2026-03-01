 # 📞Minimalistic Chat v1.7.5

Cli TCP chat based on golang
So you can download it and use now in releases. Made for x86/64🖥️
Also creates API next to it

Become my mentor please, I really like what I do!

## Architecture
<img width="715" height="456" alt="изображение" src="https://github.com/user-attachments/assets/edb6addd-badb-4286-971e-3107b8aeb891" />

## Requirements

1. (If you wanna take my code) - Go 1.22 at least

2. Friend or somebody else and a bit of time

3. The disire to spend time here

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

/Change - changes nick 

```

Server part 

```bash

/status - all necessary information

/members - check all ip addresses and their owners!

/kick usr - kick user by nick

/ban usr - ban user by IP address

## also you'll get logs with 

```


## How to use

Firstly you need to download files (host and client parts from releases) and after you need to make file work with:


### Linux

```bash

## Make your certificate on vps! (Linux only)

openssl genrsa -out ca.key 2048

openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=MyLocalCA"

sudo nano server.conf ## I use nano

## !Enter your config with your IP address in it

openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr -config server.conf

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extfile server.conf -extensions v3_ext

## on vps

./server ## or your name of file

## copy ca.crt from the place where you generated cert to your local machine

./client ## or your name of file

## thats all

```

### Windows
```bash
## All what you need to do in Linux but

./client.exe ## or your file name
```

I also want you to send your feedback about using this console app :3
I need to take a rest because I fell myself very tired due to school and limitless coding
I will return in two days I think
## Star it please it will provide me for future updates 
