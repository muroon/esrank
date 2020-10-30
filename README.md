# esrank

**A real-time score ranking system using redis**

System specifications

- Ranking is higher in descending order of score
- When same score, earlier scorer gets higher rank

Normally, when sorting by redis sorted set, members with the same score are ranked by sorting by member's key. This package is useful when you want to give a higher rank to the member who registered the score earlier.

The ranking can be displayed immediately after adding the score. So there is no need to set up a queuing system or batch for ranking.

<img width="100%" alt="esrank_rule" src="https://user-images.githubusercontent.com/301822/97300850-51b86000-189a-11eb-9965-a84fac71d6b4.png">

See [here](https://github.com/muroon/esrank/blob/master/doc/redis.md) for a detailed explanation of how esrank uses redis.

## Performance

Performance comparison when using either esrank or DB with the above specifications. Shows the execution time of [RankingList(Get List)](#get-listget-100-rankings-from-first) and [GetRanking(Get Personal Rank)](#get-personal-rank) respectively.

response time

|-|esrank(redis) (ms)|DB (ms)|esrank/DB (%)|
|---|---|---|---|
|RankingList|0.02816603333|0.4238205667|6|
|GetRanking|0.01488041333|0.1489615633|10|

As a result, esrank takes less than 10% of the processing time compared to using DB. 
There is a way to use the cache partially, but it is simpler to use esrank.

## How to use

### New Ranking

The redis client is required which is in [github.com/go-redis/redis/v8](https://github.com/go-redis/redis).

```
import(
	goredislib "github.com/go-redis/redis/v8"
	"github.com/muroon/esrank"
)
	
	# using go-redis/v8
	client = goredislib.NewClient(&goredislib.Options{
		Addr:     "localhost:6379",
	})

	# new monthly ranking
	now := time.Now()
	st := time.Date(now.Year(), now.Month(), 1, 0, 0, 0,0, time.UTC)
	rank := NewRanking(
		client, // redis client
		Name("my_monthly_ranking"), // ranking name
		StartTime(st), // start time 
	)
```

### Add Score

```
	userID := 1
	score := 100
	err := rank.AddRankingScore(ctx, userID, score)
```

### Get List(get 100 rankings from first)

```
	rankingList, err := rank.RankingList(ctx, 0, 99)
```

### Get Personal Rank

```
	rank, score, err := rank.GetRanking(ctx, userID)
```

### Examples

[Examples](https://github.com/muroon/esrank/tree/master/examples)

## Limitations

1. uid (member's id) is uint32 type
2. The expiration date of ranking depends on TimeMode

## TimeMode

The maximum elapsed time from StartTime differs depending on the TimeMode.
When TimeMode is TimeModeMilliSec, if there are multiple same scores, the difference of score registration time of 1 millisec or more is compared and sorted. Time differences less than 1 milli sec cannot be handled.
Similarly, in the case of TimeModeSec, the time difference of 1sec or more is compared, and in the case of TimeModeMicroSec, it is 1microsec.

|TimeMode|Expiration|
|---|---|
|TimeModeMicroSec|136 years|
|TimeModeMilliSec (Default)|1.36 years|
|TimeModeSec|4.97 days|

```
	rank := NewRanking(
		client,
		Name("my_monthly_ranking"),
		StartTime(st),
		SetTimeMode(TimeModeMicroSec), // timemode(microsec)
	)
```


