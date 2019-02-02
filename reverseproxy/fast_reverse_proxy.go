package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)


var proxyClient = &fasthttp.HostClient{
	Addr: "47.244.99.172:80",
	// set other options here if required - most notably timeouts.
}

func ReverseProxyHandler(ctx *fasthttp.RequestCtx) {
	requestUri:=string(ctx.Request.RequestURI());
	fmt.Printf("ReverseProxyHandler........%s \n",requestUri)
	req := &ctx.Request
	//路由映射,自定义uri映射
	//req.Header.SetRequestURI("/test.php")
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