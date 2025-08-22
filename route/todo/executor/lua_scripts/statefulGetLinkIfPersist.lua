local a=redis.call("GET",KEYS[1])if a and tonumber(redis.call("TTL",KEYS[1]))==-1 then return {a}else return {'-1'}end
