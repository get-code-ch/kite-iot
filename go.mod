module github.com/get-code-ch/kite-iot

go 1.15

replace (
 github.com/get-code-ch/kite-common => D:/projects/kite-common
 github.com/get-code-ch/ads1115 => D:/projects/ads1115
)

require (
	github.com/get-code-ch/ads1115 v0.0.0-20201030085128-84512810eb26
	github.com/get-code-ch/kite-common v0.0.0-20210101094746-7a93386692c8
	github.com/get-code-ch/mcp23008/v3 v3.0.0-20200901044239-9e2f23436931
	github.com/gorilla/websocket v1.4.2
)
