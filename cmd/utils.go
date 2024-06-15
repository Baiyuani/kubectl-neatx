package cmd

import (
	"regexp"
	s "strings"
)

func DeleteSlice3(a []string, elem string) []string {
	j := 0
	for _, v := range a {
		if v != elem {
			a[j] = v
			j++
		}
	}
	return a[:j]
}

// 删除字符串中多余的空格，有多个空格时，仅保留一个空格
func deleteExtraSpace(data string) string {
	// 替换 tab 为空格
	s1 := s.Replace(data, "\t", " ", -1)

	// 正则表达式匹配两个及两个以上空格
	regstr := "\\s{2,}"
	reg, _ := regexp.Compile(regstr)

	// 将字符串复制到切片
	s2 := make([]byte, len(s1))
	copy(s2, s1)

	// 删除多余空格
	spcIndex := reg.FindStringIndex(string(s2))
	for len(spcIndex) > 0 {
		s2 = append(s2[:spcIndex[0]+1], s2[spcIndex[1]:]...)
		spcIndex = reg.FindStringIndex(string(s2))
	}

	return string(s2)
}
