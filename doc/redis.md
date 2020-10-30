# In Redis

esrank uses stored sets of redis.
In member key, high-order 32-bit data is timestamp and low-order 32-bit data is uid(userID).

<img width="444" alt="MemberKey" src="https://user-images.githubusercontent.com/301822/97695180-fe871d00-1ae6-11eb-946e-233bf744599d.png">

So using ZREVRANGE, you can get correct ranking data, that is earlier scorer gets higher rank in same score.

<img width="816" alt="InRedis" src="https://user-images.githubusercontent.com/301822/97695277-237b9000-1ae7-11eb-873e-f7f57d0ac5aa.png">

