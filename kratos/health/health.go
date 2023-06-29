package health

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type Checker interface {
	Check(ctx context.Context) error
}
type CheckerFunc func() error

func (f CheckerFunc) Check(ctx context.Context) error {
	return f()
}

type Server struct {
	log *log.Helper
	s   *http.Server
	srv *HealthService
}

func (h *Server) Start(ctx context.Context) error {
	route := h.s.Route("/")
	route.GET("/health", func(ctx http.Context) error {
		v, err := h.srv.Health(ctx, &HealthRequest{})
		if err != nil || v.Status == Status_DOWN {
			return ctx.Result(503, v)
		}
		return ctx.Result(200, v)
	})
	return nil
}

func (h *Server) Stop(ctx context.Context) error {
	h.srv.stop <- struct{}{}
	close(h.srv.stop)
	return nil
}

func NewServer(health []Checker, logger log.Logger, s *http.Server) *Server {
	return &Server{
		s:   s,
		log: log.NewHelper(log.With(logger, "component", "health")),
		srv: NewHealthService(logger, health),
	}
}

// HealthService is a health service.
type HealthService struct {
	log        *log.Helper
	mu         sync.RWMutex
	stop       chan struct{}
	status     Status
	components map[string]Status
	checkers   []Checker
}

func NewHealthService(logger log.Logger, checkers []Checker) *HealthService {
	s := &HealthService{
		log:        log.NewHelper(logger),
		checkers:   checkers,
		status:     Status_DOWN,
		stop:       make(chan struct{}, 1),
		components: make(map[string]Status),
	}

	s.log.Debugf("health checker count %d", len(checkers))

	go s.checker()
	return s
}

func (s *HealthService) checker() {
	ticker := time.NewTicker(time.Millisecond * 100)

loop:
	for {
		select {
		case <-s.stop:
			ticker.Stop()
			s.log.Debug("stop health checker")
			break loop

		case <-ticker.C:
			ticker.Stop()
			for _, checker := range s.checkers {
				s.mu.Lock()
				name := toSnake(reflect.ValueOf(checker).Elem().Type().Name())
				if err := checker.Check(context.Background()); err != nil {
					s.components[name] = Status_DOWN
				} else {
					s.components[name] = Status_UP
				}
				s.mu.Unlock()

				s.log.Debugf("health check --> component: %s, status: %s", name, s.components[name])
			}
			ticker = time.NewTicker(time.Second * 5)
		}
	}

	s.log.Debug("health checker stopped")
}

func (s *HealthService) Health(_ context.Context, _ *HealthRequest) (*HealthReply, error) {
	s.status = Status_DOWN
	for _, v := range s.components {
		if v == Status_DOWN {
			s.status = Status_DOWN
			break
		}
		s.status = v
	}

	return &HealthReply{
		Status:     s.status,
		Components: s.components,
	}, nil
}

func toSnake(camel string) (snake string) {
	var b strings.Builder
	diff := 'a' - 'A'
	l := len(camel)
	for i, v := range camel {
		// A is 65, a is 97
		if v >= 'a' {
			b.WriteRune(v)
			continue
		}
		// v is capital letter here
		// irregard first letter
		// add underscore if last letter is capital letter
		// add underscore when previous letter is lowercase
		// add underscore when next letter is lowercase
		if (i != 0 || i == l-1) && (          // head and tail
		(i > 0 && rune(camel[i-1]) >= 'a') || // pre
			(i < l-1 && rune(camel[i+1]) >= 'a')) { //next
			b.WriteRune('_')
		}
		b.WriteRune(v + diff)
	}
	return b.String()
}
