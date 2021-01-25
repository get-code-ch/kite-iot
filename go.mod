module github.com/get-code-ch/kite-iot

go 1.15

//replace (
//	github.com/get-code-ch/mcp23008/v3 => D:/projects/mcp23008/v3
//	github.com/get-code-ch/ads1115 => D:/projects/ads1115
//	github.com/get-code-ch/kite-common => D:/projects/kite-common
//)

require (
	github.com/get-code-ch/ads1115 v0.0.0-20210111055700-659960de2946
	github.com/get-code-ch/kite-common v0.0.0-20210123103241-672e202876ff
	github.com/get-code-ch/mcp23008/v3 v3.0.0-20210111055510-732bb549de40
	github.com/gorilla/websocket v1.4.2
)
