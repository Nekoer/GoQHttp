package protocol

import (
	"GoQHttp/config"
	"net/http"
)

type Protocol interface {
	Init(w http.ResponseWriter, r *http.Request, config *config.Config)
}
