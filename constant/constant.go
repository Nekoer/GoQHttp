package constant

import (
	"GoQHttp/config"
	"GoQHttp/cqcode"
	"GoQHttp/openapi"
	"os"
)

var (
	LogFile       *os.File
	Configuration *config.Config
	OpenApi       openapi.OpenApi
	CQCode        cqcode.CQCode
)
