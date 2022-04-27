// 字符串自定义参数替换
package strReplacer

import (
	"fmt"
	"regexp"
	"strings"
)

// 替换配置
type Replace struct {
	MatchStart  string                        // 匹配开始标记
	MatchEnd    string                        // 匹配结束标记
	ParamsStart string                        // 参数开始标记
	ParamsSplit string                        // 参数分隔符号
	ParamsEnd   string                        // 参数结束标记
	params      map[string]replaceParams      // 普通参数列表
	regexParams map[string]replaceRegexParams // 正则参数列表
}

// 普通参数结构体
type replaceParams struct {
	args    bool              // 是否有参数
	handler ReplaceParamsFunc // 参数处理函数
}

// 正则参数结构体
type replaceRegexParams struct {
	args    bool                   // 是否有参数
	handler ReplaceRegexParamsFunc // 参数处理函数
}

// 普通参数处理函数
//	args - 参数解析列表
type ReplaceParamsFunc func(args ...string) string

// 正则参数处理函数
//	params - 参数名解析列表
//	args - 参数解析列表
type ReplaceRegexParamsFunc func(params []string, args ...string) string

// 默认 Replace 对象
//	r := &Replace{
//		MatchStart:  "{#",
//		MatchEnd:    "}",
//		ParamsStart: `\(`,
//		ParamsSplit: ",",
//		ParamsEnd:   `\)`,
//		params:      map[string]replaceParams{},
//		regexParams: map[string]replaceParams{},
//	}
func Default() *Replace {
	r := &Replace{
		MatchStart:  "{#",
		MatchEnd:    "}",
		ParamsStart: `\(`,
		ParamsSplit: ",",
		ParamsEnd:   `\)`,
		params:      map[string]replaceParams{},
		regexParams: map[string]replaceRegexParams{},
	}
	return r
}

// 新建 Replace 对象
//	r := &Replace{
//	 	MatchStart:  "",
//	 	MatchEnd:    "",
//	 	ParamsStart: "",
//	 	ParamsSplit: "",
//	 	ParamsEnd:   "",
//	 	params:      map[string]replaceParams{},
//	 	regexParams: map[string]replaceRegexParams{},
//	 }
func New() *Replace {
	r := &Replace{
		MatchStart:  "",
		MatchEnd:    "",
		ParamsStart: "",
		ParamsSplit: "",
		ParamsEnd:   "",
		params:      map[string]replaceParams{},
		regexParams: map[string]replaceRegexParams{},
	}
	return r
}

// AddParams - 添加普通参数
//	name - 参数名
//	handler - 参数处理函数
//	args - 是否有参数
func (r *Replace) AddParams(name string, handler func(...string) string, args bool) {
	if r.params == nil {
		r.params = make(map[string]replaceParams)
	}
	r.params[name] = replaceParams{args: args, handler: handler}
}

// AddRegexParams - 添加正则参数
//	name - 参数名是正则表达式
//	handler - 参数处理函数
//	args - 是否有参数
func (r *Replace) AddRegexParams(name string, handler func([]string, ...string) string, args bool) {
	if r.regexParams == nil {
		r.regexParams = make(map[string]replaceRegexParams)
	}
	r.regexParams[name] = replaceRegexParams{args: args, handler: handler}
}

// DelParams - 删除普通参数
//	name - 参数名
func (r *Replace) DelParams(name string) {
	delete(r.params, name)
}

// DelRegexParams - 删除正则参数
//	name - 参数名是正则表达式
func (r *Replace) DelRegexParams(name string) {
	delete(r.regexParams, name)
}

// getMatchRegex - 获取匹配正则
func (r *Replace) getMatchRegex() *regexp.Regexp {
	regStr := r.MatchStart + ".+?" + r.MatchEnd
	reg := regexp.MustCompile(regStr)
	return reg
}

// getParamsRegex - 获取参数正则
func (r *Replace) getParamsRegex() *regexp.Regexp {
	regStr := `^` + r.MatchStart + `([^` + r.ParamsStart + `]+)` + `(?:` + r.ParamsStart + `(.+?)` + r.ParamsEnd + `)?` + r.MatchEnd + `$`
	reg := regexp.MustCompile(regStr)
	return reg
}

// String - 返回替换后的字符串
//	s - 要替换的字符串
func (r *Replace) String(s string) string {
	return r.replace(s)
}

// ToString - 替换字符串
//	s - 要替换的字符串
func (r *Replace) ToString(s *string) {
	*s = r.replace(*s)
}

// replace - 替换
//	s - 要替换的字符串
func (r *Replace) replace(s string) string {
	reg := r.getMatchRegex()
	return reg.ReplaceAllStringFunc(s, r.replaceMatch)
}

// replaceMatch - 替换函数
//	s - 要替换的字符串
func (r *Replace) replaceMatch(s string) string {
	params, args := r.parseParams(s)
	fmt.Printf("参数名: %s 参数值: %v", params, args)
	if params == "" { // 没有参数
		return s
	}
	// 从正则参数中获取
	for k, v := range r.regexParams {
		reg := regexp.MustCompile(k)
		res := reg.FindStringSubmatch(params)
		if len(res) > 0 {
			if v.args { // 有参数
				if len(args) == 0 { // 没有参数值
					// fmt.Printf("参数名: %s 参数值: %v 没有获取到参数值", params, args)
					return s
				}
				return v.handler(res[1:], args...)
			}
			if len(args) != 0 { // 没有设置参数，但是有参数
				// fmt.Printf("参数名: %s 参数值: %v 没有设置参数，但是有参数", params, args)
				return s
			}
			return v.handler(res[1:])
		}
	}
	// 从普通参数中获取
	for k, v := range r.params {
		if k != params {
			continue
		}
		if v.args { // 有参数
			if len(args) == 0 { // 没有参数值
				// fmt.Printf("参数名: %s 参数值: %v 没有获取到参数值", params, args)
				return s
			}
			return v.handler(args...)
		}
		if len(args) != 0 { // 没有设置参数，但是有参数
			// fmt.Printf("参数名: %s 参数值: %v 没有设置参数，但是有参数", params, args)
			return s
		}
		return v.handler()
	}
	return s
}

// parseParams - 解析参数
//	s - 要解析的字符串
//	params - 参数名
//	args - 参数值
func (r *Replace) parseParams(s string) (params string, args []string) {
	reg := r.getParamsRegex()
	res := reg.FindStringSubmatch(s)
	if len(res) >= 2 { // 参数名
		params = res[1]
	}
	if len(res) == 3 && res[2] != "" { // 参数值
		s := strings.Split(res[2], r.ParamsSplit)
		for _, v := range s {
			if v == "" { // 有一个参数为空就不执行
				return
			}
		}
		args = s
	}
	return
}
