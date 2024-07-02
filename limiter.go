package abc

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type LimiterConsumFunc func()

// common rate limiter.
type UserRateLimiter struct {
	user   map[string]*rate.Limiter
	period map[string]int64
	per    map[string]context.Context
	mu     *sync.RWMutex
	r      rate.Limit
	b      int
}

func NewUserRateLimiter(r rate.Limit, b int) *UserRateLimiter {
	l := &UserRateLimiter{
		user:   make(map[string]*rate.Limiter),
		period: make(map[string]int64),
		per:    make(map[string]context.Context),
		mu:     &sync.RWMutex{},
		r:      r,
		b:      b,
	}
	return l
}

func (l *UserRateLimiter) AddUser(user string, period int64) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter := rate.NewLimiter(l.r, l.b)

	l.user[user] = limiter
	l.period[user] = period

	return limiter
}

func (l *UserRateLimiter) RemoveUser(user string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.user, user)
	delete(l.period, user)
	delete(l.per, user)
}

func (l *UserRateLimiter) GetLimiter(user string, period int64) (*rate.Limiter, int64) {
	l.mu.RLock()
	limiter, ok := l.user[user]
	lperiod := l.period[user]

	if !ok {
		l.mu.RUnlock()
		return l.AddUser(user, period), period
	}

	l.mu.RUnlock()

	return limiter, lperiod
}

func (l *UserRateLimiter) Todo(users ...string) context.CancelFunc {
	todo, done := context.WithCancel(context.TODO())
	l.mu.Lock()
	for _, user := range users {
		l.per[user] = todo
	}
	l.mu.Unlock()
	go func() {
		after := time.After(5 * time.Minute)
		select {
		case <-todo.Done():
			for _, user := range users {
				l.RemoveUser(user)
			}
			return
		case <-after:
		}
		return
	}()
	return done
}

func LimiterPer[T StrdNumd](queue *UserRateLimiter, user T) context.CancelFunc {
	if queue.b != 1 || queue.r != 0 {
		return func() {}
	}

	queue.mu.Lock()
	todo, ok := queue.per[ToString(user)]
	queue.mu.Unlock()
	limiter, _ := queue.GetLimiter(
		ToString(user),
		0,
	)
	defer func() {
		limiter.Allow()
	}()

	if limiter.Burst() == 0 && ok {
		select {
		case <-todo.Done():
			queue.AddUser(ToString(user), 0)
			return queue.Todo(ToString(user))
		}
	}

	if !ok {
		return queue.Todo(ToString(user))
	}

	select {
	case <-todo.Done():
		return queue.Todo(ToString(user))
	}
}

func LimiterWait[T StrdNumd](queue *UserRateLimiter, users ...T) (bool, context.CancelFunc) {
	if queue.b != 1 || queue.r != 0 {
		return false, func() {}
	}

	b := true
	l := make([]*rate.Limiter, 0)
	users2 := make([]string, 0)
	for _, user := range users {
		u := ToString(user)
		limiter, _ := queue.GetLimiter(
			u,
			0,
		)
		if limiter.Burst() == 0 {
			b = false
		}
		l = append(l, limiter)
		users2 = append(users2, u)
	}
	if !b {
		return false, func() {}
	}

	done := queue.Todo(users2...)
	for _, limiter := range l {
		limiter.Allow()
	}

	return true, done
}

func Limiter[T StrdNumd](user T, queue *UserRateLimiter, refill func() int64) (*rate.Limiter, context.CancelFunc, LimiterConsumFunc) {
	if queue.b == 0 && queue.r == 0 {
		return &rate.Limiter{}, func() {}, func() {}
	}
	todo, done := context.WithCancel(context.TODO())
	f := func() {
		select {
		case <-todo.Done():
			queue.RemoveUser(ToString(user))
			return
		default:
			done()
			limiter, period := queue.GetLimiter(
				ToString(user),
				refill(),
			)
			if time.Now().Unix() < period {
				limiter.Allow()
			}
		}
	}
	limiter, period := queue.GetLimiter(
		ToString(user),
		refill(),
	)
	if time.Now().Unix() > period {
		limiter = queue.AddUser(
			ToString(user),
			refill(),
		)
	}

	return limiter, done, func() { f() }
}
