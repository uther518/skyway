package main

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/skyhttprewrite"
	"log"
)

var proxyClient = &fasthttp.HostClient{
	Addr: "47.244.99.172:80",
	// set other options here if required - most notably timeouts.
}

/**
 * 重定向后的回调
 */
func RewritedRequest(ctx *fasthttp.RequestCtx, rewriteUri *skyhttprewrite.RewriteUri) {
	ctx.Logger().Printf("RewritedRequest原始URI:%s \n", ctx.Request.URI().Path())
	ctx.Logger().Printf("RewritedRequest重写后URI:%s \n", rewriteUri.RewriteUri)

	ctx.URI().SetPath(rewriteUri.RewriteUri)
	ctx.Request.Header.SetRequestURI(rewriteUri.RewriteUri)
	if len(rewriteUri.RewriteQueryString) > 0 {
		var buffer bytes.Buffer
		if len(ctx.URI().QueryString()) > 0 {
			buffer.Write(ctx.URI().QueryString())
			buffer.WriteByte('&')
		}
		buffer.WriteString(rewriteUri.RewriteQueryString)
		ctx.URI().SetQueryString(buffer.String())
	}

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

/**
 * 重定向
 */
func main() {

	defer func() {
		if err := recover(); err != nil {
			log.Fatal("终于捕获到了panic产生的异常：", err)
		}
	}()

	rewrite := skyhttprewrite.New()
	a := skyhttprewrite.RewriteUri{};
	a.OriginUri = "/hello/{name}/test/{foo}"
	a.DestUri = "/test.php?hello=$1&test=$2";
	b := skyhttprewrite.RewriteUri{};
	b.OriginUri = "/test/{path}"
	b.DestUri = "/hello/test/$1"

	c := skyhttprewrite.RewriteUri{};
	c.OriginUri = "/user/{id}?age={age}"
	c.DestUri = "/user/$1/age/$2"

	d := skyhttprewrite.RewriteUri{};
	d.OriginUri = "/foo/bar"
	d.DestUri = "/v1/bar/foo"

	rewrite.GET(a.OriginUri, &a)
	rewrite.GET(b.OriginUri, &b)
	rewrite.GET(c.OriginUri, &c)
	rewrite.GET(d.OriginUri, &d)

	rewrite.CallBack(RewritedRequest);
	log.Fatal(fasthttp.ListenAndServe(":8080", rewrite.Handler))
}
