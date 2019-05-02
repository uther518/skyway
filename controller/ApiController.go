package controller

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func ApiRegister(ctx *fasthttp.RequestCtx) {


	fmt.Fprint(ctx, "Welcome Register!\n")
}

