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
	stop       chan struct{}
	status     Status
	components sync.Map
	checkers   []Checker
	tick       time.Duration
}

func NewHealthService(logger log.Logger, checkers []Checker) *HealthService {
	s := &HealthService{
		log:        log.NewHelper(logger),
		checkers:   checkers,
		status:     Status_DOWN,
		stop:       make(chan struct{}, 1),
		components: sync.Map{},
		tick:       time.Second * 1,
	}

	s.log.Debugf("health checker count %d", len(checkers))

	go s.checker()

	return s
}

func (s *HealthService) checker() {
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-s.stop:
			ticker.Stop()
			s.log.Debug("stop health checker")
			return
		case <-ticker.C:
			ticker.Stop()
			for _, checker := range s.checkers {
				name := toSnake(reflect.ValueOf(checker).Elem().Type().Name())
				if err := checker.Check(context.Background()); err != nil {
					s.components.Store(name, Status_DOWN)
				} else {
					s.components.Store(name, Status_UP)
				}
			}
			ticker = time.NewTicker(s.tick)
		}
	}
}

func (s *HealthService) Health(_ context.Context, _ *HealthRequest) (*HealthReply, error) {
	s.status = Status_UP
	components := make(map[string]Status)
	s.components.Range(func(key, value any) bool {
		if value.(Status) == Status_DOWN {
			s.status = Status_DOWN
		}
		components[key.(string)] = value.(Status)
		return true
	})

	return &HealthReply{
		Status:     s.status,
		Components: components,
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
