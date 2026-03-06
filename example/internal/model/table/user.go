// ==========================================================================
// Code generated and maintained by RenHe CLI tool. DO NOT EDIT.
// ==========================================================================

package table

import (
	"example/internal/model/table/internal"
)

var User = &userTable{}

type userTable struct{}

func (*userTable) Table() string {
	return internal.UserTable
}

func (*userTable) Columns() internal.UserColumns {
	return internal.UserColumnsVar
}

func (*userTable) Idents() internal.UserIdents {
	return internal.UserIdentsVar
}
