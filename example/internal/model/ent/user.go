// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package ent

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// User is the entity for table user.
type User struct {
	bun.BaseModel `bun:"table:user"`
	Id            int64           `bun:"id,pk" json:"id"`                                                         // 用户id
	Phone         string          `bun:"phone,type:varchar(255),notnull" json:"phone"`                            // 手机号码,已注销:*#时间戳
	Password      string          `bun:"password,type:varchar(255),notnull" json:"password"`                      // 登录密码,已注销账号为空串
	Nickname      string          `bun:"nickname,type:varchar(255),notnull" json:"nickname"`                      // 昵称,已注销账号为:账号已注销
	Realname      string          `bun:"realname,type:varchar(255),notnull" json:"realname"`                      // 真实姓名,空串表示未实名
	Avatar        string          `bun:"avatar,type:varchar(255),notnull" json:"avatar"`                          // 头像文件名
	Sex           string          `bun:"sex,notnull" json:"sex"`                                                  // 性别:u代表未知,m代表男性,f代表女性,x代表X性别(少之又少)
	State         bool            `bun:"state,notnull" json:"state"`                                              // 状态
	CreatedAt     time.Time       `bun:"created_at,notnull,nullzero,default:CURRENT_TIMESTAMP" json:"created_at"` // 注册时间
	UpdatedAt     time.Time       `bun:"updated_at,notnull,nullzero,default:CURRENT_TIMESTAMP" json:"updated_at"` // 最后修改时间
	DeletedAt     *time.Time      `bun:"deleted_at,nullzero,soft_delete" json:"deleted_at"`                       // 删除/注销于
	Other         json.RawMessage `bun:"other,type:jsonb" json:"other"`                                           // 其他信息
}
