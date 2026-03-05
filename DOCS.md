# Shord discription
I wanna to present u my main repo - TCP/TLS chat based on go. 
So It's my first net project regardless of my laziness. 
After about more than month of my time, its finally done and ready to work on your VPS without any problems.

# Features
1. TLS x509 encryption with self-signed certificates which guarantees baseful safety
2. Baseful Dockerfile for server deploying
3. Good potentional of updating code
4. go.mod & go.sum are also already here
5. a lot of helpful commands are also here

# How to use
*Firstly you need to dovnload all binaries in Releases
*After go to your VPS and create self-signed cartificates or just copy all what I described in README.md
```
Before all, open 1358 and 1359 ports (I chose random ones, if you want other ones, change Dokerfile's 'EXPOSE')
./chat -addr="YOUR_VPS_IPv4_ADDRESS" -p=1358 -pRst-1359
I'll see msg about server's running
```
*So now you can connect to it with help of client
*Linux version
```
./client -addr="YOUR_VPS_IPv4_ADDRESS" -p=1358
1358 is chat's port, 1359 is https or web port
```
*Windows version
```
./client.exe -addr="YOUR_VPS_IPv4_ADDRESS" -p=1358
1358 is chat's port, 1359 is https or web port
```
