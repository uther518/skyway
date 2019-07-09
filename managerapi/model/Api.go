package model

type patternType int

type Api struct {
	/**
	 * ApiID
	 */
	ApiId int
	/**
	 *Api名称
	 */
	ApiName string
	/**
	 *Api描述信息
	 */
	ApiDescription string
	/**
	 * 服务ID
	 */
	ServiceId int
	/**
	 * 接口分组ID
	 */
	GroupId int
	/**
	 * 匹配方式：
	 * 1,通过正则表达式匹配
	 * 2,通过字符串匹配
	 */
	UriPatternType patternType
	/**
	 * 来源请求URI匹配表达式
	 */
	OriginUriPattern string
	/**
	 * 后端接口URI格式
	 */
	DestUriPattern string
}

func NewApi() *Api {
	return &Api{
		ApiId: 0,
	}
}
