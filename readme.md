# GIRC

A simple irc client in go

![girc](https://github.com/user-attachments/assets/4983fc29-ea5f-4a94-abc7-0da2bd7aae7d)


## How to run
```
# once
go install

go run .                       # uses ./config.yaml
go run . -config other.yaml
go test ./...
```

## Commands
| Command | Description |
| --- | --- |
| `/join #channel [key]` | join a channel |
| `/part [#channel] [reason]` | leave a channel (defaults to the current one) |
| `/nick newnick` | change your nickname |
| `/msg nickname message` | send a private message |
| `/names [#channel]` | list the users in a channel |
| `/quit [reason]` | quit the server |
| `/help` | show the command list |

Any other text is sent to the current channel. `Tab` cycles focus between the
message log, the user list and the input; pressing `Enter` on a user starts a
private message to them.

## Preview

![image](https://github.com/user-attachments/assets/24e43f3e-896b-4c60-b175-5fbe0d7c2f0e)

## TODO
- [ ] Connect to multiple channels
- [x] Add some cli beautification
- [ ] Manage PMs better
- [ ] Support all irc client commands
- [ ] Refactor - Channels/Events between all domains
