package controller

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"skyway/managerapi/model"
	"strconv"
)

func ApiRegister(ctx *fasthttp.RequestCtx) {
	apiName := ctx.QueryArgs().Peek("apiName")
	apiIdParam := ctx.QueryArgs().Peek("apiId")
	serviceIdParam := ctx.QueryArgs().Peek("serviceId")
	apiGroupIdParam := ctx.QueryArgs().Peek("apiGroupId")
	originUrlPattern := ctx.QueryArgs().Peek("originUrlPattern")
	destUrlPattern := ctx.QueryArgs().Peek("destUrlPattern")
	apiDescription := ctx.QueryArgs().Peek("apiDescription")

	apiId, _ := strconv.Atoi(string(apiIdParam))
	serviceId, _ := strconv.Atoi(string(serviceIdParam))
	apiGroupId, _ := strconv.Atoi(string(apiGroupIdParam))

	api := model.NewApi()
	api.ApiId = apiId
	api.ApiName = string(apiName)
	api.ServiceId = serviceId
	api.GroupId = apiGroupId
	api.OriginUriPattern = string(originUrlPattern)
	api.DestUriPattern = string(destUrlPattern)
	api.ApiDescription = string(apiDescription)

	//apiName := ctx.UserValue("apiName")
	fmt.Fprint(ctx, string(apiId))
	fmt.Fprint(ctx, string(apiName))
	fmt.Fprint(ctx, "Welcome Register!\n", )
}
