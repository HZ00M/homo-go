local a=redis.call("DEL",KEYS[1])if ARGV[2]~="-1" then redis.call("EXPIRE",KEYS[1],tonumber(ARGV[2]))end;redis.call("HDEL",KEYS[2],ARGV[1])return {a}
