package api

import (
	"errors"
	"fmt"
	"strings"
	"tomatoBlogDB/global"
	"tomatoBlogDB/model"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

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
		m.Fail(model.ResponseJson{
			Msg:  "fail to phrase parameters: " + errs.Error(),
			Code: 400,
		})
		return false
	}

	// 4. Validate
	if err := validate.Struct(option.DTO); err != nil {
		m.AddError(err)
		m.Fail(model.ResponseJson{
			Msg:  m.translateValidationError(err),
			Code: 422,
		})
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
