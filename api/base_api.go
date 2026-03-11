package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/global"
	"tomatoBlogDB/model"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

// TODO! 已完成自定义错误类型appError, Dao层必须抛出appError并立马区分BizErr和SysErr
// TODO! Service层必须抛出BizErr和SysErr, 并且需要立马区分出数据校验错误（手写、或封装）
var validate = validator.New()

type BaseApi struct {
	Ctx    iris.Context
	Errors error
	Logger *zap.SugaredLogger
}

type BuildRequestOption struct {
	Ctx       iris.Context
	DTO       any
	BindUrl   bool // whether bind URL Params
	BindQuery bool //whether bind Query String
	BindBody  bool // whether bind Body
}

func NewBaseApi() *BaseApi {
	return &BaseApi{
		Logger: global.Logger,
	}
}

func (m *BaseApi) SetContext(ctx iris.Context) *BaseApi {
	m.Ctx = ctx
	return m
}

func (m *BaseApi) AddError(errNew error) {
	// Go 1.20+ recommend
	m.Errors = errors.Join(m.Errors, errNew)
}

func (m *BaseApi) GetError() error {
	return m.Errors
}

// BuildRequest return bool, if false: fail to bind or validate, Controller should return directly
// = 40091XX bizCode to track BuildRequest()
func (m *BaseApi) BuildRequest(option BuildRequestOption) bool {
	if option.DTO == nil {
		return true
	}

	var errs error
	// 1.  Bind URL Params (RESTful path params)
	if option.BindUrl {
		if err := m.Ctx.ReadParams(option.DTO); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	// 2. Bind Query Params
	if option.BindQuery {
		if err := m.Ctx.ReadQuery(option.DTO); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	// 3. Bind Body
	if option.BindBody {
		// Iris will choose suitable format to bind Body according to Content-Type
		// JSON, XML, Form, YAML
		if err := m.Ctx.ReadBody(option.DTO); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		m.AddError(errs)
		m.HandleError(m.Ctx, errcode.NewBizErr(4009100, "fail to phrase parameters: "+errs.Error()))

		return false
	}

	// 4. Validate
	if err := validate.Struct(option.DTO); err != nil {
		m.AddError(err)
		m.HandleError(m.Ctx, errcode.NewBizErr(4009101, m.translateValidationError(err)))

		return false
	}
	return true
}

func (m *BaseApi) translateValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, fe := range ve {
			msg := fmt.Sprintf("Field[%s] validate failure: %s", fe.Field(), fe.Tag())
			errMsgs = append(errMsgs, msg)
		}
		return strings.Join(errMsgs, "; ")
	}
	return err.Error()
}

func (m *BaseApi) Fail(resp model.ResponseJson) {
	model.Fail(m.Ctx, resp)
}

func (m *BaseApi) Ok(resp model.ResponseJson) {
	model.Ok(m.Ctx, resp)
}

func (m *BaseApi) ServerFail(resp model.ResponseJson) {
	model.ServerFail(m.Ctx, resp)
}

// = Dispatch errors
func (m *BaseApi) HandleError(ctx iris.Context, err error) {
	if err == nil {
		return
	}

	var appError *errcode.AppError
	// 1. if warped as appError
	if errors.As(err, &appError) {
		if appError.HttpCode >= 500 {
			global.Logger.Errorw("System Error Caught",
				"path", ctx.Path(),
				"biz_code", appError.BizCode,
				"raw_error", appError.RawError,
			)
		} else {
			global.Logger.Infow("Business Blocked",
				"path", ctx.Path(),
				"reason", appError.Msg,
			)
		}

		m.Fail(model.ResponseJson{
			Status: appError.HttpCode, // 假设你的 Fail 方法内部会提取 Status 作为 HTTP 状态码
			Code:   appError.BizCode,
			Msg:    appError.Msg,
		})
		return
	}

	// 2.1 if validationErrors
	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		m.Fail(model.ResponseJson{
			Status: http.StatusBadRequest,
			Code:   40001,
			Msg:    "JSON parameters invalid, field [" + unmarshalTypeErr.Field + "] expects " + unmarshalTypeErr.Type.String(),
		})
		return
	}

	// 2.2 拦截 JSON 格式本身损坏的错误 (比如少了个逗号、括号没闭合)
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		m.Fail(model.ResponseJson{
			Status: http.StatusBadRequest,
			Code:   40001,
			Msg:    "JSON format invalid, please check syntax and data struct",
		})
		return
	}

	// 2.2 if validationErrors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		m.Fail(model.ResponseJson{
			Status: http.StatusBadRequest,
			Code:   40001,
			Msg:    "Invalid parameters: " + err.Error(),
		})
		return
	}

	// 兜底处理：万一某处直接 return errors.New("野生错误")
	m.Fail(model.ResponseJson{
		Status: http.StatusInternalServerError, // 假设你的 Fail 方法内部会提取 Status 作为 HTTP 状态码
		Code:   50000,
		Msg:    "服务器内部未知异常: " + err.Error(),
	})
}
