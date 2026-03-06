// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package ent

import (
	"time"

	"github.com/uptrace/bun"
)

// User is the entity for table user.
type User struct {
	bun.BaseModel `bun:"table:user"`
	Id            int64      `bun:"id,pk" json:"id"`              // 用户id
	Phone         string     `bun:"phone" json:"phone"`           // 手机号码,已注销:*#时间戳
	Password      string     `bun:"password" json:"password"`     // 登录密码,已注销账号为空串
	Nickname      string     `bun:"nickname" json:"nickname"`     // 昵称,已注销账号为:账号已注销
	Realname      string     `bun:"realname" json:"realname"`     // 真实姓名,空串表示未实名
	Avatar        string     `bun:"avatar" json:"avatar"`         // 头像文件名
	Sex           string     `bun:"sex" json:"sex"`               // 性别:u代表未知,m代表男性,f代表女性,x代表X性别(少之又少)
	State         bool       `bun:"state" json:"state"`           // 状态
	CreatedAt     time.Time  `bun:"created_at" json:"created_at"` // 注册时间
	UpdatedAt     time.Time  `bun:"updated_at" json:"updated_at"` // 最后修改时间
	DeletedAt     *time.Time `bun:"deleted_at" json:"deleted_at"` // 删除/注销于
	Other         string     `bun:"other" json:"other"`           // 其他信息
}
