package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

var proxyClient = &fasthttp.HostClient{
	Addr: "localhost:8080",
	// set other options here if required - most notably timeouts.
}

func ReverseProxyHandler(ctx *fasthttp.RequestCtx) {

	fmt.Println("ReverseProxyHandler........")


	req := &ctx.Request
	resp := &ctx.Response
	prepareRequest(req)
	if err := proxyClient.Do(req, resp); err != nil {
		ctx.Logger().Printf("error when proxying the request: %s", err)
	}
	postprocessResponse(resp)
}

func prepareRequest(req *fasthttp.Request) {
	// do not proxy "Connection" header.
	req.Header.Del("Connection")
	// strip other unneeded headers.

	// alter other request params before sending them to upstream host
}

func postprocessResponse(resp *fasthttp.Response) {
	// do not proxy "Connection" header
	resp.Header.Del("Connection")

	// strip other unneeded headers

	// alter other response data if needed
}

func main() {
	if err := fasthttp.ListenAndServe(":80", ReverseProxyHandler); err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}