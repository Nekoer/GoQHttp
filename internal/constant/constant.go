package constant

import (
	"GoQHttp/config"
	"GoQHttp/internal/cqcode"
	"GoQHttp/internal/openapi"
	"os"
)

var (
	LogFile       *os.File
	Configuration *config.Config
	OpenApi       openapi.OpenApi
	CQCode        cqcode.CQCode
)
