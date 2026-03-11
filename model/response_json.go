package model

import (
	"reflect"

	"github.com/kataras/iris/v12"
)

type ResponseJson struct {
	Status int    `json:"status,omitempty"` // 0 means not set, use default status code
	Code   int    `json:"code,omitempty"`   // used to indicate error to front-end engineer
	Msg    string `json:"msg,omitempty"`    // message to front-end user
	Data   any    `json:"data,omitempty"`
	Total  int64  `json:"total,omitempty"`
}

func buildStatus(resp ResponseJson, nDefaultStatus int) int {
	if resp.Status == 0 {
		return nDefaultStatus
	}
	return resp.Status
}

func (m ResponseJson) IsEmpty() bool {
	return reflect.DeepEqual(m, ResponseJson{})
}

func Ok(ctx iris.Context, resp ResponseJson) {
	HttpResponse(ctx, buildStatus(resp, 200), resp)
	// = alternative
	// ctx.StopWithJSON(buildStatus(resp, 200), resp)
}

func Fail(ctx iris.Context, resp ResponseJson) {
	HttpResponse(ctx, buildStatus(resp, 400), resp)
	// = alternative
	// ctx.StopWithJSON(buildStatus(resp, 400), resp)
}

func ServerFail(ctx iris.Context, resp ResponseJson) {
	HttpResponse(ctx, buildStatus(resp, 500), resp)
}

func HttpResponse(ctx iris.Context, status int, resp ResponseJson) {
	if resp.IsEmpty() {
		ctx.StopWithJSON(status, nil)
		return
	} else {
		ctx.StopWithJSON(status, resp)
	}
}
