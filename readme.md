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

## Features
- Several channels and private conversations at once, each in its own buffer,
  plus a status buffer for server messages
- Buffer list with unread counts; private messages and mentions of your nick
  are highlighted
- Per-channel user list with operator/voice prefixes, kept up to date on
  joins, parts, quits, nick and mode changes
- Channel topics shown in the buffer title
- Replies to CTCP VERSION, PING, TIME and CLIENTINFO
- TLS support (`ssl: true` in `config.yaml`)

## Commands
| Command | Description |
| --- | --- |
| `/join #channel[,#other] [key]` | join channels |
| `/part [#channel] [reason]` | leave a channel |
| `/close [reason]` | close the buffer, leaving its channel |
| `/query nick [message]` | open a private conversation |
| `/msg target message` | send a message |
| `/me action` | send an action |
| `/notice target message` | send a notice |
| `/ctcp nick command [args]` | send a CTCP request |
| `/nick newnick` | change your nickname |
| `/topic [new topic]` | show or set the channel topic |
| `/mode [target] [modes [params]]` | show or change modes |
| `/op`, `/deop`, `/voice`, `/devoice nick...` | change channel privileges |
| `/ban nick\|mask...`, `/unban mask...` | manage bans |
| `/kick nick [reason]` | kick a user from the channel |
| `/invite nick [#channel]` | invite a user to a channel |
| `/whois nick`, `/who mask` | look up users |
| `/away [message]` | set, or clear without a message, your away status |
| `/list [pattern]`, `/names [#channel]` | list channels / users |
| `/quote command` | send a raw command to the server |
| `/clear` | clear the buffer |
| `/quit [reason]` | quit the server |
| `/help` | show the command list |

Any other text is sent to the current channel or conversation (start it with
`//` to send a message beginning with `/`).

## Keys
| Key | Action |
| --- | --- |
| `Ctrl+N` / `Ctrl+P` | next / previous buffer |
| `Alt+1` ... `Alt+0` | jump to buffer 1-10 |
| `PgUp` / `PgDn` | scroll the message log |
| `Tab` | move focus between input, messages, users and buffers |
| `Enter` on a user | open a private conversation |
| `Ctrl+C` | quit |

## Architecture
- `irc/` - protocol: message parsing/formatting, CTCP, the connection
- `config/` - configuration loading
- `ui/` - the tview terminal interface
- `app/` - the client: a single event loop goroutine owns all session state
  and receives server messages and user input over channels, so no locking is
  needed; its output is calls to the connection and to the UI

## Preview

![image](https://github.com/user-attachments/assets/24e43f3e-896b-4c60-b175-5fbe0d7c2f0e)

## TODO
- [x] Connect to multiple channels
- [x] Add some cli beautification
- [x] Manage PMs better
- [x] Support all irc client commands
- [x] Refactor - Channels/Events between all domains
