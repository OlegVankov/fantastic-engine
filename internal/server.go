package internal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/OlegVankov/fantastic-engine/internal/handler"
	"github.com/OlegVankov/fantastic-engine/internal/handler/accrual"
	"github.com/OlegVankov/fantastic-engine/internal/repository/postgres"
)

type Server struct {
	srv     *http.Server
	handler *handler.Handler
}

func NewServer(addr, dsn string) *Server {
	h := &handler.Handler{
		Repository: postgres.NewUserRepository(dsn),
	}

	router := chi.NewRouter()

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		r.Route("/", func(r chi.Router) {
			r.Use(handler.Auth)

			r.Post("/orders", h.Orders)
			r.Get("/orders", h.GetOrders)

			r.Post("/balance/withdraw", h.Withdraw)
			r.Get("/balance", h.Balance)
			r.Get("/withdrawals", h.Withdrawals)
		})
	})

	return &Server{
		srv: &http.Server{
			Addr:           addr,
			Handler:        router,
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
		},
		handler: h,
	}
}

func (s *Server) Run(accrualAddr string) {
	go accrual.SendAccrual(accrualAddr, s.handler)
	s.srv.ListenAndServe()
}
