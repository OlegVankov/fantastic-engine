package internal

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	go func() {
		err := s.srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server ListenAndServe", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c

	log.Println("server", "Graceful shutdown starter with signal", sig.String())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal("server", err)
	}
	log.Println("server gracefully shutdown complete")
}
