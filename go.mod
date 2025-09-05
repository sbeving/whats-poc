module github.com/kevzy/whatsappspoofing

go 1.21

toolchain go1.22.4

require (
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mdp/qrterminal/v3 v3.0.0
	go.mau.fi/whatsmeow v0.0.0-20230805111647-405414b9b5c0
	google.golang.org/protobuf v1.33.0
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	go.mau.fi/libsignal v0.1.0 // indirect
	go.mau.fi/util v0.4.1 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	rsc.io/qr v0.2.0 // indirect
)

replace github.com/kevzy/whatsappspoofing => ./

replace go.mau.fi/whatsmeow => ./whatsmeow
