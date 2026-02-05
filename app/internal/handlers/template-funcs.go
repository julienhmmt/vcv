package handlers

import "html/template"

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add": func(args ...int) int {
			sum := 0
			for _, v := range args {
				sum += v
			}
			return sum
		},
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
}
