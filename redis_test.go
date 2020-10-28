package esrank

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var client *goredislib.Client

func connectRedisClient(ctx context.Context) error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	if client == nil {
		client = goredislib.NewClient(&goredislib.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

func newClientForTest(ctx context.Context) (*Ranking, error) {
	if err := connectRedisClient(ctx); err != nil {
		return nil, err
	}
	rank := NewRanking(client)
	return rank, nil
}

func flushAllForTest(ctx context.Context) error {
	return client.FlushAll(ctx).Err()
}

func TestNewRanking(t *testing.T) {

	defaultStartTime, _ := time.Parse(TimeLayout, StartTimeDefault)

	testRankingName := "TestRanking"
	testStartTime := time.Now()

	type args struct {
		cl   *goredislib.Client
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *Ranking
	}{
		{
			name: "Default",
			args: args{
				cl: client,
			},
			want: &Ranking{
				name: RankingNameDefault,
				mode: TimeModeMilliSec,
				startTime: defaultStartTime,
				client: client,
				rs: redsync.New(goredis.NewPool(client)),
			},
		},
		{
			name: "SetRankingName",
			args: args{
				cl: client,
				opts: []Option{
					Name(testRankingName),
				},
			},
			want: &Ranking{
				name: testRankingName,
				mode: TimeModeMilliSec,
				startTime: defaultStartTime,
				client: client,
				rs: redsync.New(goredis.NewPool(client)),
			},
		},
		{
			name: "SetStartTime",
			args: args{
				cl: client,
				opts: []Option{
					StartTime(testStartTime),
				},
			},
			want: &Ranking{
				name: RankingNameDefault,
				mode: TimeModeMilliSec,
				startTime: testStartTime,
				client: client,
				rs: redsync.New(goredis.NewPool(client)),
			},
		},
		{
			name: "TimeModeMicro",
			args: args{
				cl: client,
				opts: []Option{
					SetTimeMode(TimeModeMicroSec),
				},
			},
			want: &Ranking{
				name: RankingNameDefault,
				mode: TimeModeMicroSec,
				startTime: defaultStartTime,
				client: client,
				rs: redsync.New(goredis.NewPool(client)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRanking(tt.args.cl, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRanking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddRankingScore(t *testing.T) {
	ctx := context.Background()

	rank, err := newClientForTest(ctx)
	if err != nil {
		t.Error(err)
	}

	if err := flushAllForTest(ctx); err != nil {
		t.Error(err)
	}

	type args struct {
		uid   uint32
		score float64
		rank int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRank int64
		wantScore float64
	}{
		{
			name: "uid:1",
			args: args{
				uid: 1,
				score: 10,
			},
			wantRank: 0,
			wantScore: 50, // 10 + 40 (add score)
		},
		{
			name: "uid:2",
			args: args{
				uid: 2,
				score: 20,
			},
			wantRank: 3, // same score higher rank
			wantScore: 20,
		},
		{
			name: "uid:3",
			args: args{
				uid: 3,
				score: 30,
			},
			wantRank: 2,
			wantScore: 30,
		},
		{
			name: "uid:4",
			args: args{
				uid: 4,
				score: 40,
			},
			wantRank: 1,
			wantScore: 40,
		},
		{
			name: "uid:5",
			args: args{
				uid: 5,
				score: 20,
			},
			wantRank: 4, // same score lower rank
			wantScore: 20,
		},
	}
	for _, tt := range tests {
		if err := rank.AddRankingScore(ctx, tt.args.uid, tt.args.score); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Millisecond)
	}
	if err := rank.AddRankingScore(ctx, 1, 40); err != nil {
		t.Error(err)
	}

	// check GetRanking
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank, sc, err := rank.GetRanking(ctx, tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRankingScore() uid:%d, error = %v, wantErr %v", tt.args.uid, err, tt.wantErr)
			}
			if rank != tt.wantRank {
				t.Errorf("AddRankingScore() uid:%d, rank: %v, correct rank: %v", tt.args.uid, rank, tt.wantRank)
			}
			if sc != tt.wantScore {
				t.Errorf("AddRankingScore() uid:%d, score: %v, correct score: %v", tt.args.uid, sc, tt.wantScore)
			}
		})
	}

	// check RankingList
	last := len(tests) - 1
	rankingList, err := rank.RankingList(ctx, 0, int64(last))
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("RankingList_%s", tt.name), func(t *testing.T) {
			personal := rankingList[tt.wantRank]
			var score float64
			var ok bool
			if score, ok = personal[tt.args.uid]; !ok {
				t.Errorf("RankingList() uid:%d, rank: %v, personal: %v", tt.args.uid, tt.wantRank, personal)
			}
			if score != tt.wantScore {
				t.Errorf("RankingList() uid:%d, rank: %v, score: %v, correctScore:%v", tt.args.uid, tt.wantRank, score, tt.wantScore)
			}
		})
	}
}

func TestAddRankingScoreMulti(t *testing.T) {
	ctx := context.Background()

	rank, err := newClientForTest(ctx)
	if err != nil {
		t.Error(err)
	}

	uid := uint32(1)
	cnt := uint32(5)
	score := uint32(10)

	if err := flushAllForTest(ctx); err != nil {
		t.Error(err)
	}

	added := make(chan bool, cnt)
	for i := uint32(0); i < cnt; i++ {
		go func(uid, i uint32) {
			if err := rank.AddRankingScore(ctx, uid, float64(score)); err != nil {
				t.Error(err)
			}
			added <- true
		}(uid, i)
	}

	for i := uint32(0); i < cnt; i++ {
		switch <-added {
		}
	}

	type args struct {
		uid   uint32
		score int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want uint32
	}{
		{
			name: "",
			args: args{
				uid: uid,
				score: int64(score),
			},
			want: cnt * score,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sc, err := rank.GetRanking(ctx, tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRankingScore() error = %v, wantErr %v", err, tt.wantErr)
			}
			if uint32(sc) != tt.want {
				t.Errorf("AddRankingScore() score: %v, correct score: %v", sc, tt.want)
			}
		})
	}
}

func TestRanking_Remove(t *testing.T) {
	ctx := context.Background()

	rank, err := newClientForTest(ctx)
	if err != nil {
		t.Error(err)
	}

	if err := flushAllForTest(ctx); err != nil {
		t.Error(err)
	}

	targetUID := uint32(3)

	type args struct {
		uid   uint32
		score float64
		rank int64
	}
	argList := []args{
		args{
			uid: 1,
			score: 10,
		},
		args{
			uid: 2,
			score: 20,
		},
		args{
			uid: 3,
			score: 30,
		},
		args{
			uid: 4,
			score: 40,
		},
		args{
			uid: 5,
			score: 50,
		},
	}
	for _, a := range argList {
		if err := rank.AddRankingScore(ctx, a.uid, a.score); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Millisecond)
	}

	if err = rank.Remove(ctx, targetUID); err != nil {
		t.Error(err)
	}

	rankingList, err := rank.RankingList(ctx, 0, int64(len(argList)))
	if err != nil {
		t.Error(err)
	}

	for i, ranking := range rankingList {
		if score, ok := ranking[targetUID]; ok {
			t.Errorf("Remove error. rank:%d, uid:%d, score:%v", i, targetUID, score)
		}
	}
}

func TestRanking_RemoveAll(t *testing.T) {
	ctx := context.Background()
	if err := connectRedisClient(ctx); err != nil {
		t.Error(err)
	}

	if err := flushAllForTest(ctx); err != nil {
		t.Error(err)
	}

	rankingName := "removeRank"

	rank := NewRanking(client, Name(rankingName))

	if err := rank.RemoveAll(ctx); err != nil {
		t.Error(err)
	}

	uids := []uint32{1, 2}
	score := float64(100)
	for _, uid := range uids {
		if err := rank.AddRankingScore(ctx, uid, score); err != nil {
			t.Error(err)
		}
	}

	v, err := client.Exists(ctx, rank.rankingListName()).Result()
	if err != nil {
		t.Error(err)
	}

	if v == 0 {
		t.Errorf("targetRanking is none. %v", rank.rankingListName())
	}

	keys, err := client.Keys(ctx, rank.uidKeys()).Result()
	if err != nil {
		t.Error(err)
	}
	if len(keys) != len(uids) {
		t.Errorf("invalid data. keys:%#+v", keys)
	}

	if err := rank.RemoveAll(ctx); err != nil {
		t.Error(err)
	}

	v, err = client.Exists(ctx, rank.rankingListName()).Result()
	if err != nil {
		t.Error(err)
	}

	if v != 0 {
		t.Errorf("targetRanking exists. %v", rank.rankingListName())
	}

	keys, err = client.Keys(ctx, rank.uidKeys()).Result()
	if err != nil {
		t.Error(err)
	}
	if len(keys) > 0 {
		t.Errorf("invalid data after RemoveAll. keys:%#+v", keys)
	}
}