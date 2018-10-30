package service

import (
	"fmt"
)

const todoIDParam = "id"

var todoRootURL = "/todos"
var todoItemURL = fmt.Sprintf("%s/{%s}", todoRootURL, todoIDParam)
