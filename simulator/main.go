package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	orderv1 "market/proto/orderservice/v1"
	spotv1 "market/proto/spotinstrumentservice/v1"
	userv1 "market/proto/userservice/v1"
)

const registerTimeout = 20 * time.Second

type stats struct {
	created, cancelled, alreadyDone atomic.Int64
	mu                              sync.Mutex
	errs                            map[codes.Code]int64
}

func (s *stats) fail(err error) {
	s.mu.Lock()
	s.errs[status.Code(err)]++
	s.mu.Unlock()
}

func (s *stats) snapshotErrs() map[codes.Code]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make(map[codes.Code]int64, len(s.errs))
	for k, v := range s.errs {
		cp[k] = v
	}
	return cp
}

// market хранит общую «справедливую цену» (блуждает случайно).
type market struct {
	mu   sync.RWMutex
	fair float64
}

func (m *market) price() float64 { m.mu.RLock(); defer m.mu.RUnlock(); return m.fair }

func (m *market) drift() {
	m.mu.Lock()
	m.fair *= 1 + (rand.Float64()-0.5)*0.004 // ±0.2% за шаг
	m.mu.Unlock()
}

type bot struct {
	ctx context.Context
	mu  sync.Mutex
	own []string // id открытых ордеров, чтобы было что отменять
}

func (b *bot) pushOrder(id string) {
	b.mu.Lock()
	b.own = append(b.own, id)
	b.mu.Unlock()
}

func (b *bot) popOrder() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.own) == 0 {
		return "", false
	}
	id := b.own[0]
	b.own = b.own[1:]
	return id, true
}

func dial(target string) *grpc.ClientConn {
	c, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return c
}

func main() {
	bots := flag.Int("bots", 5, "number of traders")
	rate := flag.Float64("rps", 5, "total orders per second")
	dur := flag.Duration("duration", 0, "0 = until Ctrl+C")
	cancelProb := flag.Float64("cancel", 0.1, "probability to cancel an own order instead of placing")
	spread := flag.Float64("spread", 0.02, "price spread around fair price (0.02 = ±1%)")
	reqTimeout := flag.Duration("timeout", 3*time.Second, "per-request timeout")
	maxInflight := flag.Int("inflight", 50, "max concurrent requests")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *dur > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *dur)
		defer cancel()
	}

	users := userv1.NewUserServiceClient(dial("127.0.0.1:50051"))
	spot := spotv1.NewSpotInstrumentServiceClient(dial("127.0.0.1:50052"))
	orders := orderv1.NewOrderServiceClient(dial("127.0.0.1:50053"))

	// регистрируем ботов
	pool := make([]*bot, 0, *bots)
	for i := 0; i < *bots; i++ {
		email := fmt.Sprintf("bot%d-%d@sim.test", i, time.Now().UnixNano())
		rctx, cancel := context.WithTimeout(ctx, registerTimeout)
		start := time.Now()
		resp, err := users.Register(rctx, &userv1.RegisterRequest{
			Email:     email,
			Password:  "Passw0rd!sim",
			FirstName: fmt.Sprintf("Bot%d", i),
			LastName:  "Sim",
		}) // VERIFY
		cancel()
		if err != nil {
			fmt.Println("register:", err)
			return
		}
		fmt.Printf("bot %d registered in %s\n", i, time.Since(start))
		token := resp.GetAuthToken().GetAccessToken() // VERIFY
		pool = append(pool, &bot{ctx: metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)})
	}

	// инструмент и стартовая цена
	lctx, lcancel := context.WithTimeout(pool[0].ctx, *reqTimeout)
	list, err := spot.ListInstruments(lctx, &spotv1.ListInstrumentsRequest{}) // VERIFY
	lcancel()
	if err != nil {
		fmt.Println("list instruments:", err)
		return
	}
	var instr string
	fair := 100.0
	for _, in := range list.GetInstruments() {
		if in.GetStatus() == spotv1.InstrumentStatus_INSTRUMENT_STATUS_ACTIVE { // VERIFY
			instr = in.GetId()
			if v, err := strconv.ParseFloat(in.GetCurrentRate(), 64); err == nil && v > 0 {
				fair = v
			}
			break
		}
	}
	if instr == "" {
		fmt.Println("no ACTIVE instrument")
		return
	}
	mkt := &market{fair: fair}
	fmt.Printf("instrument=%s start=%.2f bots=%d rps=%.1f spread=±%.1f%% timeout=%s\n",
		instr, fair, *bots, *rate, *spread*50, *reqTimeout)

	st := &stats{errs: map[codes.Code]int64{}}

	// статистика раз в секунду + дрейф цены
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				mkt.drift()
				fmt.Printf("fair=%.2f created=%d cancelled=%d already_done=%d errors=%v\n",
					mkt.price(), st.created.Load(), st.cancelled.Load(), st.alreadyDone.Load(), st.snapshotErrs())
			}
		}
	}()

	interval := time.Duration(float64(time.Second) / *rate)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	var wg sync.WaitGroup
	sem := make(chan struct{}, *maxInflight)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick.C:
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break loop
		}
		wg.Add(1)
		b := pool[rand.Intn(len(pool))]
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			rctx, cancel := context.WithTimeout(b.ctx, *reqTimeout)
			defer cancel()

			if rand.Float64() < *cancelProb {
				if id, ok := b.popOrder(); ok {
					_, err := orders.CancelOrder(rctx, &orderv1.CancelOrderRequest{Id: id}) // VERIFY
					switch {
					case err == nil:
						st.cancelled.Add(1)
					case status.Code(err) == codes.FailedPrecondition:
						// ордер уже исполнен/отменён: ожидаемый отказ
						st.alreadyDone.Add(1)
					default:
						st.fail(err)
					}
					return
				}
			}

			side := orderv1.OrderSide_ORDER_SIDE_BUY
			if rand.Intn(2) == 0 {
				side = orderv1.OrderSide_ORDER_SIDE_SELL
			}
			price := fmt.Sprintf("%.2f", mkt.price()*(1+(rand.Float64()-0.5)**spread))
			qty := fmt.Sprintf("%d", 1+rand.Intn(5))
			resp, err := orders.CreateOrder(rctx, &orderv1.CreateOrderRequest{
				InstrumentId: instr, Side: side, Type: orderv1.OrderType_ORDER_TYPE_LIMIT, // VERIFY
				Price: &price, Quantity: qty,
			})
			if err != nil {
				st.fail(err)
				return
			}
			st.created.Add(1)
			if resp.GetOrder().GetStatus() == orderv1.OrderStatus_ORDER_STATUS_OPEN { // VERIFY
				b.pushOrder(resp.GetOrder().GetId())
			}
		}()
	}
	wg.Wait()
	fmt.Println("stopped")
}
