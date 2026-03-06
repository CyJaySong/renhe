// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package do

// User is the data operation struct for table user.
type User struct {
	Id        any // 用户id
	Phone     any // 手机号码,已注销:*#时间戳
	Password  any // 登录密码,已注销账号为空串
	Nickname  any // 昵称,已注销账号为:账号已注销
	Realname  any // 真实姓名,空串表示未实名
	Avatar    any // 头像文件名
	Sex       any // 性别:u代表未知,m代表男性,f代表女性,x代表X性别(少之又少)
	State     any // 状态
	CreatedAt any // 注册时间
	UpdatedAt any // 最后修改时间
	DeletedAt any // 删除/注销于
	Other     any // 其他信息
}
