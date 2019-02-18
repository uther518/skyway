package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/skyhttprewrite"
	"log"
	"regexp"
)

// Index is the index handler
func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

// Hello is the Hello handler
func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

func HelloTest(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "HelloTest, %s!\n", ctx.UserValue("name"))
}

// MultiParams is the multi params handler
func MultiParams(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hi, %s, %s!\n", ctx.UserValue("name"), ctx.UserValue("word"))
}

// QueryArgs is used for uri query args test #11:
// if the req uri is /ping?name=foo, output: Pong! foo
// if the req uri is /piNg?name=foo, redirect to /ping, output: Pong!
func QueryArgs(ctx *fasthttp.RequestCtx) {
	name := ctx.QueryArgs().Peek("name")
	fmt.Fprintf(ctx, "Pong! %s\n", string(name))
}

func RegExpUri() {
	//reg = regexp.MustCompile(`(Hello)(.*)(Go)`)
	uri := "/hello/foo1111/test/name2222"
	regstr := "/hello/{foo}/test/{name}"

	regstr1 := `(\{\w+\})`
	reg1 := regexp.MustCompile(regstr1);
	regstr = reg1.ReplaceAllString(regstr, `(\w+)`);
	fmt.Printf("=============%q\n", regstr)
	reg := regexp.MustCompile(regstr);
	fmt.Printf("%q\n", reg.ReplaceAllString(uri, "/test/${1}/hello/${2}"))
	fmt.Printf("%q\n", reg.ReplaceAllString(uri, "/test/hello.php?name=$1&age=$2"))
}

/**
 * 重定向
 */
func main() {
	rewrite := skyhttprewrite.New()
	a := skyhttprewrite.RewriteUri{};
	a.OriginUri="/hello/{name}/test/{foo}"
	a.DestUri = "/test.php?hello=$1&test=$2";

	b := skyhttprewrite.RewriteUri{};
	b.OriginUri="/test/{path}"
	b.DestUri = "/hello/test/$1"

	c := skyhttprewrite.RewriteUri{};
	c.OriginUri="/user/{id}?age={g}"
	c.DestUri = "/user/$1/age/${age}"

	rewrite.GET(a.OriginUri, &a)
	rewrite.GET(b.OriginUri, &b)
	//rewrite.GET(c.OriginUri, &c)

	log.Fatal(fasthttp.ListenAndServe(":8080", rewrite.Handler))
}
