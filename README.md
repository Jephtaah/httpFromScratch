# httpFromScratch

An HTTP/1.1 server built directly on top of raw TCP sockets in Go — no `net/http` on the server side, no shortcuts.

## Why this exists

Every web framework you've ever used is hiding something from you: HTTP is just text over a TCP connection. `net/http` parses that text, manages the connections, and hands you a nice `Request`/`ResponseWriter` pair so you never have to think about any of it.

This project rips that abstraction away. I wanted to actually understand what happens between a client opening a socket and a server sending back `200 OK` — byte by byte — instead of trusting that the framework "just handles it." So I wrote a server that listens on a raw TCP socket, reads the incoming bytes myself, parses the request line and headers by hand, and writes the response back out manually.

## What it does

- Listens on a TCP port using Go's `net` package — nothing higher-level
- Parses the request line (method, path, HTTP version) directly from the byte stream
- Parses headers key by key, no shortcuts
- Reads the request body using `Content-Length`
- Builds and writes HTTP responses manually — status line, headers, body, all assembled by hand
- Supports `GET` and `POST`
- Routes requests to handlers (`GET /health` → `200 OK`, `POST /echo` → echoes the body back)
- Handles multiple connections concurrently, one goroutine per connection
- Implements keep-alive vs. connection-close behavior properly
- Includes a manual DNS query/resolution routine, so I could see what resolution actually looks like at the byte level, not just call `net.LookupHost` and move on

## What it doesn't do

- Import `net/http` anywhere in the server itself. It only shows up in the test client used to hit the server from the outside — which felt like fair game, since that's testing the server, not building it.
- Use any third-party HTTP or routing libraries.

## Running it

```bash
git clone https://github.com/yourusername/httpFromScratch.git
cd httpFromScratch
go run main.go
```

The server starts listening on `localhost:8080` by default.

Try it:

```bash
curl http://localhost:8080/health
curl -X POST -d "hello" http://localhost:8080/echo
```

## Project structure

```
httpFromScratch/
├── main.go              # entry point, starts the listener
├── server/
│   ├── server.go         # connection accept loop, goroutine-per-connection
│   ├── request.go        # request line + header parsing
│   ├── response.go        # response building/writing
│   └── router.go          # route matching and handler dispatch
├── dns/
│   └── resolver.go        # manual DNS query construction and parsing
├── client/
│   └── testclient.go       # net/http-based client used only for testing
└── server_test.go
```

## Things that were harder than expected

- **Header parsing.** It looks trivial until you actually sit down and handle folded headers, case-insensitivity, and the fact that the stream doesn't hand you nice clean lines — you're reading raw bytes and deciding where a line ends yourself.
- **Keep-alive.** Deciding when to close a connection versus keep it open for the next request, and making sure I wasn't leaking goroutines or hanging on a read, took a few rewrites to get right.
- **The DNS piece.** Building a query packet by hand and parsing the response made the abstraction "DNS resolves a hostname to an IP" feel a lot less magical and a lot more mechanical.

## What I'd add next

- Chunked transfer encoding
- Basic HTTP/1.1 pipelining
- A router that supports path parameters (`/user/:id`)

## What this proves

I can explain — and demonstrate — exactly what a framework like Express or Gin is doing under the hood: parsing a byte stream into a request, managing concurrent connections safely, and writing a valid response back out over the same socket. Nothing here is copied from a library. It's all hand-rolled, on purpose.