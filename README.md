# Minimalistic Chat v1.1
Classical TCP chat made with GoLang.

I have spent just 4 days for first version
IDK how much time for first update but at least 5 days of laziness
So you can download it and use now.

## Requirements
1. Golang 1.19 at least
2. Friend or somebody else and a bit of time
3. Knowleges in docker & Ansible for hosting somewhere
4. The disire to support me(it's my first MVP useful project)

## Description
The minimum-interface tcp chat. Just pet-project, nothing more.

## Useful commands
```bash
/health - sends you informatin about server
/list - sends you all active members
/help - shows all info
/exit - leave the server
/quit - leave the sever
```

## How to use
Firstly you need to download files(host and client parts)and after you need to make file work with:

### Linux
```bash
go build -o server.go
go build -o client.go

./server
// in other terminal
./client
```
### Windows
```bash
go build -o server.go
go build -o client.go

./server.exe
// in other terminal
./client.exe
```

I also want you to send your feedback about using this console app :3

## Star it please it will provide me for future updates and growing up
