package skyrewrite

import (
	"github.com/valyala/fasthttp"
	"log"
	"regexp"
	"strings"
	"sync"
)

type SkyRewrite struct {
	//---/hello/foo1111/test/name2222
	ApiId                    int    //所属API ID
	OriginUri                string //---/hello/{name}/test/{foo} uri参数表达式,用户设定
	RouterPath               string //---/hello/:name/test/:foo 路由匹配,fastrouter
	OriginReg                string //---/hello/(\w+)/test/(\w+)
	DestUri                  string //---/test/$1/hello/$2   目标uri转换
	RewriteUri               string
	RewriteQueryString       string
	Regexp                   *regexp.Regexp
	IsMatchOriginQueryString bool
	IsMatchDestQueryString   bool
	QueryParams              []string
	handle                   fasthttp.RequestHandler //回调用户处理方法
}

func New() *SkyRewrite {
	return &SkyRewrite{
		ApiId: 0,
	}
}

type RewriteHandler func(ctx *fasthttp.RequestCtx, rewrite *SkyRewrite)

var instance *SkyRewrite
var once sync.Once

func ApiRouterInstance() *SkyRewrite {
	once.Do(func() {
		instance = &SkyRewrite{}
	})
	return instance
}

/**
 {foo}形参转为正则表达式(\\w+)
 {foo}形参转为正则表达式:foo
 */
func (api *SkyRewrite) MakeRegexp() {

	//带形参匹配
	pos := strings.Index(api.OriginUri, "?")
	log.Print("params", api.OriginUri)
	if pos > 0 {
		queryStr := api.OriginUri[pos+1 : len(api.OriginUri)]
		api.OriginUri = api.OriginUri[0:pos]

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
			api.IsMatchOriginQueryString = true
			api.QueryParams = append(api.QueryParams, pairs[0])
		}
	}

	if (strings.Contains(api.DestUri, "?")) {
		api.IsMatchDestQueryString = true
	}

	regstr := `(\{\w+\})`
	regCompile := regexp.MustCompile(regstr)
	api.OriginReg = regCompile.ReplaceAllString(api.OriginUri, `(\w+)`)
	//r.DestReg = regCompile.ReplaceAllString(r.OriginUri, )
	api.Regexp = regexp.MustCompile(api.OriginReg);
	regpath := `\{(\w+)\}`
	pathCompile := regexp.MustCompile(regpath)
	api.RouterPath = pathCompile.ReplaceAllString(api.OriginUri, ":${1}")
}
