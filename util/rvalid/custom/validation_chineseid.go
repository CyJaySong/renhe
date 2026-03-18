package custom

import (
	"regexp"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var chineseIDRegex = regexp.MustCompile(`^\d{17}[\dXx]$`)

// chineseIDWeights 身份证号码前17位的加权因子。
var chineseIDWeights = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

// chineseIDCheckMap 校验码映射表，余数 0-10 分别对应 1,0,X,9,8,7,6,5,4,3,2。
var chineseIDCheckMap = [11]byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

// RegisterChineseIDValidation 向验证器注册 chineseid 标签及其中文翻译。
func RegisterChineseIDValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("chineseid", validateChineseID); err != nil {
		return
	}
	err = v.RegisterTranslation("chineseid", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("chineseid", "{0}必须为有效的身份证号码", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("chineseid", fe.Field())
			return t
		},
	)
	return
}

// validateChineseID 校验18位中国居民身份证号码。
// 规则：
//  1. 长度18位，前17位为数字，第18位为数字或X/x
//  2. 第7-14位为合法日期（YYYYMMDD），且不超过当前日期
//  3. 第18位校验码符合 ISO 7064:1983 MOD 11-2 算法
func validateChineseID(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) != 18 {
		return false
	}
	if !chineseIDRegex.MatchString(s) {
		return false
	}

	// 校验出生日期
	dateStr := s[6:14]
	birth, err := time.Parse("20060102", dateStr)
	if err != nil {
		return false
	}
	if birth.After(time.Now()) {
		return false
	}

	// 校验码计算
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(s[i]-'0') * chineseIDWeights[i]
	}
	expected := chineseIDCheckMap[sum%11]
	actual := s[17]
	if actual == 'x' {
		actual = 'X'
	}
	return actual == expected
}
