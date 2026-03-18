package custom

import (
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
)

var onceTestRegisterChineseIDValidation sync.Once

func TestRegisterChineseIDValidation(t *testing.T) {
	initValidator()
	onceTestRegisterChineseIDValidation.Do(func() {
		if err := RegisterChineseIDValidation(validate, zhTranslator); err != nil {
			t.Errorf("unexpected register: %s", err)
		}
	})
}

func TestValidateChineseID(t *testing.T) {
	TestRegisterChineseIDValidation(t)
	type testStruct struct {
		ID string `v:"chineseid"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		// 合法身份证号（校验码正确）
		{"valid_male", "11010519491231002X", false},
		{"valid_female", "110105194912310046", false},
		{"valid_lower_x", "11010519491231002x", false},
		{"valid_digit_check", "320121198501010011", false},

		// 空值
		{"empty", "", true},

		// 长度不对
		{"too_short_15", "110105491231002", true},
		{"too_short_17", "11010519491231002", true},
		{"too_long_19", "11010519491231002X1", true},

		// 非法字符
		{"has_letter_in_body", "1101051949123100aX", true},
		{"has_space", "110105 19491231002X", true},
		{"last_char_invalid", "11010519491231002A", true},

		// 非法日期
		{"invalid_month_00", "110105194900310021", true},
		{"invalid_month_13", "110105194913310021", true},
		{"invalid_day_00", "110105194912000021", true},
		{"invalid_day_32", "110105194912320021", true},
		{"invalid_date_feb30", "110105199402300024", true},

		// 未来日期
		{"future_date", "110105209912310023", true},

		// 校验码错误
		{"wrong_check_digit", "110105194912310020", true},
		{"wrong_check_x", "110105194912310021", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{ID: tt.val}
			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("chineseid(%q) err=%v, wantErr=%v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestChineseIDTranslation(t *testing.T) {
	TestRegisterChineseIDValidation(t)

	type testStruct struct {
		ID string `v:"chineseid"`
	}

	s := testStruct{ID: "abc"}
	err := validate.Struct(s)
	if err == nil {
		t.Fatal("expected error")
	}
	errs := err.(validator.ValidationErrors)
	msg := errs[0].Translate(zhTranslator)
	if !strings.Contains(msg, "身份证") {
		t.Errorf("unexpected translation: %s", msg)
	}
}
