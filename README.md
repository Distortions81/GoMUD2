# GoMUD2

GoMUD2 is a modern MUD (Multi-User Dungeon) server written in Go. It aims to be small, fast and easy to hack on while supporting a wide range of clients.

## Features

- **Threaded networking** with flood protection and asynchronous input buffering
- **Telnet option negotiation** including MTTS and NAWS
- **Unicode aware** with character map negotiation and translation
- **ANSI color with 256-colour support**
- **Secure accounts** with hashed passphrases and optional TLS/SSL
- **Chat channels, tells and emotes** with offline message storage
- **Basic world and OLC** to create and edit rooms from within the game
- **Emoji shortcuts, FIGlet fonts** and other fun extras
- Built-in help system, stats command and more

## Getting Started

1. Install **Go 1.22** or newer from [go.dev](https://go.dev/dl/).
2. Build and run the server:

   ```bash
   go build
   ./goMUD2     # or: go run main.go
   ```
3. For encrypted connections generate test certificates:

   ```bash
   ./makeTestCert.sh
   ```

   The server listens on port `7777` (plain) and `7778` (TLS) by default.

Connect with any standard MUD or telnet client and explore!

## Contributing

Pull requests and bug reports are welcome. The project is intentionally minimal so you can add your own ideas easily.

## License

This project is released under the [MIT License](LICENSE).
