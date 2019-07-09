package main

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"log"
	"skyway/gateway/skyrewrite"
	"skyway/gateway/skyrouter"
	"time"
)

var proxyClient = &fasthttp.HostClient{
	Addr: "47.244.99.172:80",
	// set other options here if required - most notably timeouts.
}

func prepareRequest(req *fasthttp.Request) {

	log.Println("prepareRequest.....")
	// do not proxy "Connection" header.
	req.Header.Del("Connection")
	// strip other unneeded headers.

	// alter other request params before sending them to upstream host
}

func postprocessResponse(resp *fasthttp.Response) {

	log.Println("postprocessResponse.....")
	// do not proxy "Connection" header
	resp.Header.Del("Connection")

	// strip other unneeded headers

	// alter other response data if needed
}

func RouterRequest(ctx *fasthttp.RequestCtx, rewriteUri *skyrewrite.SkyRewrite) {
	ctx.Logger().Printf("RewritedRequest原始URI:%s %s \n", ctx.Request.URI().Path(), ctx.Request.String())
	ctx.Logger().Printf("RewritedRequest重写后URI:%s \n", rewriteUri.RewriteUri)

	//重写URI
	ctx.URI().SetPath(rewriteUri.RewriteUri)
	ctx.Request.Header.SetRequestURI(rewriteUri.RewriteUri)

	//重写QueryString
	if len(rewriteUri.RewriteQueryString) > 0 {
		var buffer bytes.Buffer
		if len(ctx.URI().QueryString()) > 0 {
			buffer.Write(ctx.URI().QueryString())
			buffer.WriteByte('&')
		}
		buffer.WriteString(rewriteUri.RewriteQueryString)
		ctx.URI().SetQueryString(buffer.String())
	}

	start := time.Now()
	req := &ctx.Request
	resp := &ctx.Response
	prepareRequest(req)
	if err := proxyClient.Do(req, resp); err != nil {
		ctx.Logger().Printf("error when proxying the request: %s", err)
	}

	postprocessResponse(resp)
	cost := time.Since(start).Nanoseconds() / 1e6
	ctx.Logger().Printf("Response Cost:%d MS,Status=%d,[%s],\n", cost, resp.StatusCode(), resp.Header.Header())
}

func main() {
	router := skyrouter.New()
	a := skyrewrite.New()
	a.OriginUri = "/hello/{name}/test/{foo}"
	a.DestUri = "/test.php?hello=$1&test=$2"
	a.ApiId = 1000

	b := skyrewrite.New()
	b.OriginUri = "/test/{path}"
	b.DestUri = "/hello/test/$1"
	b.ApiId = 1001

	c := skyrewrite.New()
	c.OriginUri = "/user/{id}/{age}?addr={addr}"
	c.DestUri = "/user/$1/age/$2/addr/$3"
	c.ApiId = 1003

	d := skyrewrite.New()
	d.OriginUri = "/foo/bar"
	d.DestUri = "/v1/bar/foo"
	d.ApiId = 1004

	router.GET(a.OriginUri, a)
	router.GET(b.OriginUri, b)
	router.GET(c.OriginUri, c)
	router.GET(d.OriginUri, d)

	router.RewriteHandle(RouterRequest)

	httpServer := fasthttp.Server{
		Handler: router.Handler,
		Name:    "skyway",
	}

	if err := httpServer.ListenAndServe(":888"); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
