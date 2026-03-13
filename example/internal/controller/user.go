package controller

import (
	"context"
	"github.com/uptrace/bun"
	"net/http"

	"example/internal/model/ent"

	"github.com/cyjaysong/renhe/database/rdb"
	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/labstack/echo/v4"
)

type User struct{}

type UserListReq struct {
	rhttp.HttpApiMeta `path:"/user" method:"GET" name:"用户列表"`
}

type UserListRes struct {
	List []ent.User `json:"list"`
}

func (u *User) List(ctx echo.Context, req *UserListReq) (*UserListRes, error) {
	var users []ent.User
	db := rdb.Database()
	err := db.NewSelect(ctx.Request().Context()).Model(&users).Scan(ctx.Request().Context())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return &UserListRes{List: users}, nil
}

type UserGetReq struct {
	rhttp.HttpApiMeta `path:"/user/:id" method:"GET" name:"用户详情"`
	Id                int64 `param:"id"`
}

type UserGetRes struct {
	User ent.User `json:"user"`
}

func (u *User) Get(ctx echo.Context, req *UserGetReq) (*UserGetRes, error) {
	var user ent.User
	db := rdb.Database()
	err := db.NewSelect(ctx.Request().Context()).Model(&user).Where("id = ?", req.Id).Scan(ctx.Request().Context())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return &UserGetRes{User: user}, nil
}

type UserCreateReq struct {
	rhttp.HttpApiMeta `path:"/user" method:"POST" name:"创建用户"`
	Phone             string `json:"phone"`
	Nickname          string `json:"nickname"`
}

type UserCreateRes struct {
	User ent.User `json:"user"`
}

func (u *User) Create(ctx echo.Context, req *UserCreateReq) (*UserCreateRes, error) {
	user := ent.User{
		Phone:    req.Phone,
		Nickname: req.Nickname,
	}
	db := rdb.Database()
	_, err := db.NewInsert(ctx.Request().Context()).Model(&user).Exec(ctx.Request().Context())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return &UserCreateRes{User: user}, nil
}

type UserTxReq struct {
	rhttp.HttpApiMeta `path:"/user/tx-demo" method:"POST" name:"事务示例"`
	Phone             string `json:"phone"`
	Nickname          string `json:"nickname"`
}

type UserTxRes struct {
	User ent.User `json:"user"`
}

func (u *User) TxDemo(ctx echo.Context, req *UserTxReq) (*UserTxRes, error) {
	db := rdb.Database()
	var user ent.User

	err := db.RunInTx(ctx.Request().Context(), nil, func(txCtx context.Context, tx bun.Tx) error {
		user = ent.User{
			Phone:    req.Phone,
			Nickname: req.Nickname,
		}
		_, err := db.NewInsert(txCtx).Model(&user).Exec(txCtx)
		if err != nil {
			return err
		}
		_, err = db.NewUpdate(txCtx).Model(&user).Set("nickname = ?", req.Nickname+" (via tx)").Where("id = ?", user.Id).Exec(txCtx)
		return err
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return &UserTxRes{User: user}, nil
}
