package custom

import (
	"github.com/go-playground/validator/v10"
	"strings"
	"sync"
	"testing"
)

var onceTestRegisterChineseValidation sync.Once

func TestRegisterChineseValidation(t *testing.T) {
	initValidator()
	onceTestRegisterChineseValidation.Do(func() {
		if err := RegisterChineseValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestValidateChinese(t *testing.T) {
	TestRegisterChineseValidation(t)

	type testStruct struct {
		Name string `v:"chinese"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"normal_name", "张三", false},
		{"three_chars", "李小明", false},
		{"four_chars", "司马相如", false},
		{"minority_dot", "买买提·艾力", false},
		{"minority_bullet", "买买提•艾力", false},
		{"single_char", "张", true},
		{"empty", "", true},
		{"leading_dot", "·张三", true},
		{"trailing_dot", "张三·", true},
		{"consecutive_dots", "买买··提", true},
		{"has_digit", "张三1", true},
		{"has_letter", "张三a", true},
		{"has_space", "张 三", true},
		{"too_long", strings.Repeat("啊", 21), true},
		{"max_len", strings.Repeat("啊", 20), false},
		{"all_ascii", "zhangsan", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Name: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("chinese(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestChineseTranslation(t *testing.T) {
	TestRegisterChineseValidation(t)

	type testStruct struct {
		Name string `v:"chinese"`
	}

	s := testStruct{Name: "a"}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "中文姓名") {
		t.Errorf("unexpected translation: %s", msg)
	}
}
