package public

import (
	"github.com/ridewindx/mel"
)

type Server struct {
	*mel.Mel
}

func NewServer() *Server {
	srv := &Server{
		Mel: mel.New(),
	}

	srv.Get("/", func(c *mel.Context) {
		c.Query()
	})
}
