// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package fasthttprouter is a trie based high performance HTTP request router.
//
// A trivial example is:
//
// package main

// import (
//     "fmt"
//     "log"
//
//     "github.com/buaazp/fasthttprouter"
//     "github.com/valyala/fasthttp"
// )

// func Index(ctx *fasthttp.RequestCtx) {
//     fmt.Fprint(ctx, "Welcome!\n")
// }

// func Hello(ctx *fasthttp.RequestCtx) {
//     fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
// }

// func main() {
//     router := fasthttprouter.New()
//     router.GET("/", Index)
//     router.GET("/hello/:name", Hello)

//     log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
// }
//
// The router matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods GET, POST, PUT, PATCH and DELETE shortcut functions exist to
// register handles, for all other methods router.Handle can be used.
//
// The registered path, against which the router matches incoming requests, can
// contain two types of parameters:
//  Syntax    Type
//  :name     named parameter
//  *name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//  Path: /blog/:category/:post
//
//  Requests:
//   /blog/go/request-routers            match: category="go", post="request-routers"
//   /blog/go/request-routers/           no match, but the router would redirect
//   /blog/go/                           no match
//   /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all parameters must always be the final path element.
//  Path: /files/*filepath
//
//  Requests:
//   /files/                             match: filepath="/"
//   /files/LICENSE                      match: filepath="/LICENSE"
//   /files/templates/article.html       match: filepath="/templates/article.html"
//   /files                              no match, but the router would redirect
//
// The value of parameters is inside ctx.UserValue
// To retrieve the value of a parameter:
//  // use the name of the parameter
//  user := ps.UserValue("user")
//

package skyhttprewrite

import (
	"github.com/prometheus/common/log"
	"github.com/valyala/fasthttp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	defaultContentType = []byte("text/plain; charset=utf-8")
	questionMark       = []byte("?")
)

type RewriteUri struct {
	//---/hello/foo1111/test/name2222
	OriginUri          string //---/hello/{name}/test/{foo} uri参数表达式,用户设定
	RouterPath         string //---/hello/:name/test/:foo 路由匹配,fastrouter
	DestUri            string //---/test/$1/hello/$2   目标uri转换
	OriginReg          string //---/hello/(\w+)/test/(\w+)
	RewriteUri         string
	RewriteQueryString string
	Regexp             *regexp.Regexp
	IsMatchOriginQueryString bool
	IsMatchDestQueryString bool
	QueryParams        []string
	handle             fasthttp.RequestHandler //回调用户处理方法
}

/**
 {foo}形参转为正则表达式(\\w+)
 {foo}形参转为正则表达式:foo
 */
func (r *RewriteUri) MakeRegexp() {

	//带形参匹配
	pos := strings.Index(r.OriginUri, "?")
	log.Info("params", r.OriginUri)
	if pos > 0 {
		queryStr := r.OriginUri[pos+1 : len(r.OriginUri)]
		r.OriginUri = r.OriginUri[0:pos]

		params := strings.Split(queryStr, "&")
		for _, pair := range params {
			pairs := strings.Split(pair, "=")

			index := strings.IndexByte(pairs[1], '{')
			if index == 0 {
				index = strings.IndexByte(pairs[1], '}')
				if index <= 0 {
					continue
				}
			} else if index < 0 {
				index := strings.IndexByte(pairs[1], ':')
				if index != 0 {
					continue
				}
			} else {
				continue
			}
			r.IsMatchOriginQueryString = true
			r.QueryParams = append(r.QueryParams, pairs[0])
		}
	}

	if(strings.Contains(r.DestUri,"?")){
		r.IsMatchDestQueryString=true;
	}

	regstr := `(\{\w+\})`
	regCompile := regexp.MustCompile(regstr)
	r.OriginReg = regCompile.ReplaceAllString(r.OriginUri, `(\w+)`)
	//r.DestReg = regCompile.ReplaceAllString(r.OriginUri, )
	r.Regexp = regexp.MustCompile(r.OriginReg);
	regpath := `\{(\w+)\}`
	pathCompile := regexp.MustCompile(regpath)
	r.RouterPath = pathCompile.ReplaceAllString(r.OriginUri, ":${1}")
}

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Rewrite struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound fasthttp.RequestHandler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed fasthttp.RequestHandler

	CallBackHandle RewriteRequestHandler //回调用户处理方法

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(*fasthttp.RequestCtx, interface{})
}

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Rewrite {
	return &Rewrite{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		CallBackHandle:         nil,
	}
}

type RewriteRequestHandler func(ctx *fasthttp.RequestCtx,rewriteUri *RewriteUri)


func (r *Rewrite) CallBack(handler RewriteRequestHandler) {
	r.CallBackHandle = handler
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Rewrite) GET(path string, rewriteUri *RewriteUri) {
	r.Handle("GET", path, rewriteUri)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Rewrite) HEAD(path string, rewriteUri *RewriteUri) {
	r.Handle("HEAD", path, rewriteUri)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Rewrite) OPTIONS(path string, rewriteUri *RewriteUri) {
	r.Handle("OPTIONS", path, rewriteUri)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Rewrite) POST(path string, rewriteUri *RewriteUri) {
	r.Handle("POST", path, rewriteUri)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Rewrite) PUT(path string, rewriteUri *RewriteUri) {
	r.Handle("PUT", path, rewriteUri)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Rewrite) PATCH(path string, rewriteUri *RewriteUri) {
	r.Handle("PATCH", path, rewriteUri)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Rewrite) DELETE(path string, rewriteUri *RewriteUri) {
	r.Handle("DELETE", path, rewriteUri)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Rewrite) Handle(method, path string, rewriteUri *RewriteUri) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	rewriteUri.OriginUri = path
	rewriteUri.MakeRegexp()
	path = rewriteUri.RouterPath;
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, rewriteUri)
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
//     router.ServeFiles("/src/*filepath", "/var/www")
/*
func (r *Router) ServeFiles(path string, rootPath string) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}
	prefix := path[:len(path)-10]

	fileHandler := fasthttp.FSHandler(rootPath, strings.Count(prefix, "/"))

	r.GET(path, func(ctx *fasthttp.RequestCtx) {
		fileHandler(ctx)
	})
}*/
func (r *Rewrite) recv(ctx *fasthttp.RequestCtx) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(ctx, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Rewrite) Lookup(method, path string, ctx *fasthttp.RequestCtx) (*RewriteUri, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path, ctx)
	}
	return nil, false
}

func (r *Rewrite) allowed(path, reqMethod string) (allow string) {
	if path == "*" || path == "/*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _ := r.trees[method].getValue(path, nil)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

/**
 *根据重写规则，重写请求
 */
func (r *Rewrite) RewriteRequest(ctx *fasthttp.RequestCtx, rewriteUri *RewriteUri) {
	destUri := rewriteUri.DestUri
	//传统api转为restful api
	if rewriteUri.IsMatchOriginQueryString {
		values, _ := url.ParseQuery(string(ctx.URI().QueryString()))
		for _, key := range rewriteUri.QueryParams {
			val := values.Get(key)
			idx := ctx.UserValueSize() + 1
			sep := "$" + strconv.Itoa(idx)
			destUri = strings.Replace(destUri, sep, val, -1)
		}
	}
	uriPath := string(ctx.URI().Path())
	rewriteUri.RewriteUri=rewriteUri.Regexp.ReplaceAllString(uriPath, destUri)

	if rewriteUri.IsMatchDestQueryString {
		querys:=strings.Split(rewriteUri.RewriteUri,"?");
		rewriteUri.RewriteUri=querys[0]
		rewriteUri.RewriteQueryString=querys[1]
	}

	//fmt.Fprintf(ctx, "%q\n", rewriteUri.Regexp.ReplaceAllString(uriPath, destUri))
	r.CallBackHandle(ctx, rewriteUri);
}

/**
 TODO 001::每个请求过来后会调用此路由方法
 */
// Handler makes the router implement the fasthttp.ListenAndServe interface.
func (r *Rewrite) Handler(ctx *fasthttp.RequestCtx) {
	if r.PanicHandler != nil {
		defer r.recv(ctx)
	}

	path := string(ctx.URI().PathOriginal())
	queryString := string(ctx.URI().QueryString())

	log.Info("skyhttprouter Handler start:", path, queryString)
	method := string(ctx.Method())
	if root := r.trees[method]; root != nil {
		if writeUri, tsr := root.getValue(path, ctx); writeUri != nil {
			r.RewriteRequest(ctx, writeUri)
			return
		} else if method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}
			log.Info("skyhttprouter Handler tsr:", tsr)

			if tsr && r.RedirectTrailingSlash {
				var uri string
				if len(path) > 1 && path[len(path)-1] == '/' {
					uri = path[:len(path)-1]
				} else {
					uri = path + "/"
				}

				if len(ctx.URI().QueryString()) > 0 {
					uri += "?" + string(ctx.QueryArgs().QueryString())
				}

				ctx.Redirect(uri, code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)

				if found {
					queryBuf := ctx.URI().QueryString()
					if len(queryBuf) > 0 {
						fixedPath = append(fixedPath, questionMark...)
						fixedPath = append(fixedPath, queryBuf...)
					}
					uri := string(fixedPath)
					ctx.Redirect(uri, code)
					return
				}
			}
		}
	}

	if method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, method); len(allow) > 0 {
				ctx.Response.Header.Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, method); len(allow) > 0 {
				ctx.Response.Header.Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed(ctx)
				} else {
					ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
					ctx.SetContentTypeBytes(defaultContentType)
					ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
				}
				return
			}
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound(ctx)
	} else {
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
			fasthttp.StatusNotFound)
	}
}
