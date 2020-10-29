package examples

import (
	"fmt"
	"time"
	"context"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/muroon/esrank"
)

func Example_monthlyRankingExample() {
	ctx := context.Background()
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:6379",
	})

	// new monthly ranking
	now := time.Now()
	st := time.Date(now.Year(), now.Month(), 1, 0, 0, 0,0, time.Local)
	rank := esrank.NewRanking(
		client, // redis client
		esrank.Name(fmt.Sprintf("my_monthly_ranking_%d_%d", st.Year(), st.Month())), // ranking name
		esrank.StartTime(st), // start time
		esrank.SetTimeMode(esrank.TimeModeMilliSec), // monthly ranking can use mill sec and sec modes.
	)

	// remove last month ranking
	lastSt := st.AddDate(0, -1, 0)
	if err := esrank.NewRanking(client,
		esrank.Name(fmt.Sprintf("my_monthly_ranking_%d_%d", lastSt.Year(), lastSt.Month())), // ranking name
	); err != nil {
		panic(err)
	}

	// add score
	if err := rank.AddRankingScore(ctx, 1, 40); err != nil {
		panic(err)
	}

	// get ranking list
	list, err := rank.RankingList(ctx, 0, 100)
	if err != nil {
		panic(err)
	}

	for i, r := range list {
		for uid, score := range r {
			fmt.Printf("%d: %v: %v\n", i, uid, score)
		}
	}

	// get personal rank and score
	ra, score, err := rank.GetRanking(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("uid:%d, rank:%d, score:%v\n", 1, ra, score )
}

func Example_dailyRankingExample() {
	ctx := context.Background()
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:6379",
	})

	// new daily ranking
	now := time.Now()
	st := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0,0, time.Local)
	rank := esrank.NewRanking(
		client,
		esrank.Name(fmt.Sprintf("my_daily_ranking_%d_%d_%d", st.Year(), st.Month(), st.Day())), // ranking name
		esrank.StartTime(st),
		esrank.SetTimeMode(esrank.TimeModeMicroSec), // daily ranking can all time modes includes micro sec mode.
	)

	// remove last daily ranking
	lastSt := st.AddDate(0, 0, -1)
	if err := esrank.NewRanking(client,
		esrank.Name(fmt.Sprintf("my_daily_ranking_%d_%d_%d", lastSt.Year(), lastSt.Month(), lastSt.Day())), // ranking name
	); err != nil {
		panic(err)
	}

	// add score
	if err := rank.AddRankingScore(ctx, 1, 40); err != nil {
		panic(err)
	}
}

func Example_yearlyRankingExample() {
	ctx := context.Background()
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:6379",
	})

	// new yearly ranking
	now := time.Now()
	st := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0,0, time.Local)
	rank := esrank.NewRanking(
		client,
		esrank.Name(fmt.Sprintf("my_yearly_ranking_%d", st.Year())), // ranking name
		esrank.StartTime(st),
		esrank.SetTimeMode(esrank.TimeModeSec), // yearly ranking can use nothing without sec mode.
	)

	// remove last yearly ranking
	lastSt := st.AddDate(0, 0, -1)
	if err := esrank.NewRanking(client,
		esrank.Name(fmt.Sprintf("my_yearly_ranking_%d", lastSt.Year())), // ranking name
	); err != nil {
		panic(err)
	}

	// add score
	if err := rank.AddRankingScore(ctx, 1, 40); err != nil {
		panic(err)
	}
}
