# SSHChat
a simple chat server written in go. designed to server a chat ui over the ssh protocol.

to run the server simply build the binary "go build -o dist/", then run dist/SSHChat.

to connect to the server simply run "ssh <user>@<server> -p <port> -t <room>"
  
The ssh server uses public key authentication, should you receive a "no publickey presented" error, please run ssh-keygen and follow the necessary steps.
