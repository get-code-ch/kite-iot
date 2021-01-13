module github.com/get-code-ch/kite-iot

go 1.15

replace (
	github.com/get-code-ch/mcp23008/v3 => D:/projects/mcp23008/v3
	github.com/get-code-ch/ads1115 => D:/projects/ads1115
)

require (
	github.com/get-code-ch/ads1115 v0.0.0-20210103171030-03ed0d150f8d
	github.com/get-code-ch/kite-common v0.0.0-20210109173656-2140e459491e
	github.com/get-code-ch/mcp23008/v3 v3.0.0-00010101000000-000000000000
	github.com/gorilla/websocket v1.4.2
	periph.io/x/conn/v3 v3.6.7 // indirect
)
