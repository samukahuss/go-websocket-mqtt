module github.com/samukahuss/go-websocket-mqtt/client

go 1.24.5

require (
	github.com/gorilla/websocket v1.5.3
	github.com/samukahuss/go-websocket-mqtt/backend v0.0.0-20251122191253-9c806e064152
	google.golang.org/protobuf v1.36.10
)

replace github.com/samukahuss/go-websocket-mqtt/client/internal/proto => ../client/internal/proto

replace github.com/samukahuss/go-websocket-mqtt/client => ../client
