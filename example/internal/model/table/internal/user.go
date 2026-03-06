// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import "github.com/uptrace/bun"

const UserTable = "user"

// UserColumns defines the column names for table user.
type UserColumns struct {
	Id        string // 用户id
	Phone     string // 手机号码,已注销:*#时间戳
	Password  string // 登录密码,已注销账号为空串
	Nickname  string // 昵称,已注销账号为:账号已注销
	Realname  string // 真实姓名,空串表示未实名
	Avatar    string // 头像文件名
	Sex       string // 性别:u代表未知,m代表男性,f代表女性,x代表X性别(少之又少)
	State     string // 状态
	CreatedAt string // 注册时间
	UpdatedAt string // 最后修改时间
	DeletedAt string // 删除/注销于
	Other     string // 其他信息
}

// UserIdents defines the column ident for table user.
type UserIdents struct {
	Id        bun.Ident // 用户id
	Phone     bun.Ident // 手机号码,已注销:*#时间戳
	Password  bun.Ident // 登录密码,已注销账号为空串
	Nickname  bun.Ident // 昵称,已注销账号为:账号已注销
	Realname  bun.Ident // 真实姓名,空串表示未实名
	Avatar    bun.Ident // 头像文件名
	Sex       bun.Ident // 性别:u代表未知,m代表男性,f代表女性,x代表X性别(少之又少)
	State     bun.Ident // 状态
	CreatedAt bun.Ident // 注册时间
	UpdatedAt bun.Ident // 最后修改时间
	DeletedAt bun.Ident // 删除/注销于
	Other     bun.Ident // 其他信息
}

var UserColumnsVar = UserColumns{
	Id:        "id",
	Phone:     "phone",
	Password:  "password",
	Nickname:  "nickname",
	Realname:  "realname",
	Avatar:    "avatar",
	Sex:       "sex",
	State:     "state",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
	Other:     "other",
}

var UserIdentsVar = UserIdents{
	Id:        bun.Ident("id"),
	Phone:     bun.Ident("phone"),
	Password:  bun.Ident("password"),
	Nickname:  bun.Ident("nickname"),
	Realname:  bun.Ident("realname"),
	Avatar:    bun.Ident("avatar"),
	Sex:       bun.Ident("sex"),
	State:     bun.Ident("state"),
	CreatedAt: bun.Ident("created_at"),
	UpdatedAt: bun.Ident("updated_at"),
	DeletedAt: bun.Ident("deleted_at"),
	Other:     bun.Ident("other"),
}
