package esrank

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

const TimeLayout string = "2006-01-01"
const StartTimeDefault string = "2020-01-01"
const RankingListNameSuffix string = "ranking"
const RankingNameDefault string = "myRanking"

type TimeMode int
const (
	TimeModeMicroSec TimeMode = 0
	TimeModeMilliSec TimeMode = 1
	TimeModeSec TimeMode = 2
)

// Ranking struct of ranking
type Ranking struct {
	name string
	mode TimeMode
	startTime time.Time
	client *goredislib.Client
	rs *redsync.Redsync
}

// NewRanking constructor function
func NewRanking(cl *goredislib.Client, opts... Option) *Ranking {
	r := new(Ranking)
	r.client = cl
	pool := goredis.NewPool(r.client)
	r.rs = redsync.New(pool)
	r.name = RankingNameDefault
	r.mode = TimeModeMilliSec
	r.startTime, _ = time.Parse(TimeLayout, StartTimeDefault)

	// options
	for _, opt := range opts {
		r = opt(r)
	}

	return r
}

// AddRankingScore add score
func (r *Ranking) AddRankingScore(ctx context.Context, uid uint32, score float64) error {
	// lock
	mutex := r.rs.NewMutex(r.lockKey(uid), redsync.WithExpiry(time.Second))
	for {
		if lockErr := mutex.Lock(); lockErr == nil {
			break
		}
	}

	// Release the lock so other processes or threads can obtain a lock.
	defer func() {
		_, _ = mutex.Unlock()
	}()

	pKey := r.uidKey(uid)
	oldKey, _ := r.getRankingKey(ctx, uid)
	var oldScore float64
	var err error
	if oldKey != "" {
		oldScore, err = r.client.ZScore(ctx,r.rankingListName(), oldKey).Result()
		if err != nil {
			return err
		}
	}

	newKey := r.setRankingKey(uid)
	mem := &goredislib.Z {
		Score: score + oldScore,
		Member: newKey,
	}

	err = r.client.ZAdd(ctx,r.rankingListName(), mem).Err()
	if err != nil {
		return err
	}

	if err = r.client.Set(ctx, pKey, newKey, 0).Err(); err != nil {
		return err
	}

	if oldKey != "" && oldKey != newKey {
		err = r.client.ZRem(ctx,r.rankingListName(), oldKey).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// RankingList ranking list of uid a score sets
func (r *Ranking) RankingList(ctx context.Context, start, end int64) ([]map[uint32]float64, error) {
	res, err := r.client.ZRevRangeWithScores(ctx,r.rankingListName(), start, end ).Result()
	if err != nil {
		return nil, err
	}

	list := make([]map[uint32]float64, 0, len(res))

	for _, re := range res {
		list = append(list, map[uint32]float64{getRankingUID(re.Member.(string)): re.Score})
	}
	return list, err
}

// GetRanking get personal rank and score
func (r *Ranking) GetRanking(ctx context.Context, uid uint32) (int64, float64, error) {
	key, err := r.getRankingKey(ctx, uid)
	if err != nil {
		return 0, float64(0), err
	}
	if key == "" {
		return 0, float64(0), nil
	}

	rank, err := r.client.ZRevRank(ctx, r.rankingListName(), key).Result()
	if err != nil {
		return 0, float64(0), err
	}

	score, err := r.client.ZScore(ctx, r.rankingListName(), key).Result()
	return rank, score, err
}

func (r *Ranking) getUnixTimeStamp(t time.Time) int64 {
	switch r.mode {
	case TimeModeMicroSec:
		return t.UnixNano()
	case TimeModeMilliSec:
		return t.UnixNano()/int64(time.Millisecond)
	case TimeModeSec:
		return t.Unix()
	}
	return 0
}

func (r *Ranking) setRankingKey(uid uint32) string {
	timestamp := r.getUnixTimeStamp(time.Now()) // int64
	t1 := uint64(timestamp) - uint64(r.getUnixTimeStamp(r.startTime))
	t2 := uint64(^t1)
	t3 := t2 << 32
	v := t3 + uint64(uid)
	return fmt.Sprintf("%v", v)
}

func getRankingUID(key string) uint32 {
	keyNum, _ := strconv.ParseUint(key, 10, 64)
	uidMax := uint64(math.MaxUint32) // 32bit max
	return uint32(keyNum & uidMax)
}

func (r *Ranking) getRankingKey(ctx context.Context, uid uint32) (string, error) {
	val, err := r.client.Exists(ctx, r.uidKey(uid)).Result()
	if err != nil {
		return "", err
	}

	if val == 0 {
		return "", nil
	}

	return r.client.Get(ctx, r.uidKey(uid)).Result()
}

func (r *Ranking) lockKey(uid interface{}) string {
	return fmt.Sprintf("esrank_%s_lock_%v", r.name, uid)
}

func (r *Ranking) uidKey(uid interface{}) string {
	return fmt.Sprintf("esrank_%s_uid_%v", r.name, uid)
}

func (r *Ranking) rankingListName() string {
	return fmt.Sprintf("esrank_%s_%s", r.name, RankingListNameSuffix)
}
