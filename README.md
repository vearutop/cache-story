# cache-story

This is a demo application to showcase a few things about in-memory caching and explain design decisions
for [github.com/bool64/cache](https://github.com/bool64/cache).

_TL;DR_ In-memory caching is a great way to improve performance and resiliency of an application at cost of memory and
delayed data consistency. You need to take care of concurrent updates, error caching, failover handling, background
updates, expiration jitter and cache warmup with transfer.

## The Story

Caching is one of the most efficient techniques to improve performance, because the fastest way to get rid of a task is
skipping it. Unfortunately caching is not a silver bullet, in some cases you can not afford reusing result of a task
due to transactionality/consistency constraints. Cache invalidation is notoriously one
of [two hard things](https://martinfowler.com/bliki/TwoHardThings.html) in Computer Science.

It is best when domain operates on immutable data and so cache invalidation is not necessary. In such case cache is
usually a net benefit. However, if there are requirements to keep mutable data in sync, cache invalidation is necessary.
The simplest strategy is to invalidate cache based on time to live (TTL). Even if seems like a bad fit compared to
event-based invalidation, consider simplicity and portability. Events do not guarantee timely delivery, in worst case
scenarios (for example if event broker is temporary down or overloaded) events could be even less precise than TTL.

Short TTL is often a good compromise between performance and consistency. It would reduce the load under heavy traffic
acting as a barrier to the data source. For the low traffic impact would be negligible.

### Demo Application

Let's start with a simple demo application. It will receive URL with query parameters and respond with a JSON object
determined by those parameters. Unique results will be stored in database to make things realistically slow.

We're going to put some load on the application with
a [custom](https://github.com/vearutop/cache-story/blob/master/cmd/cplt/cplt.go)
[`plt`](https://github.com/vearutop/plt).

Custom `plt` has additional parameters:

* `cardinality` - number of unique URLs to be generated, this affects cache hit rate,
* `group` - number of requests with similar URL being sent at once, this imitates concurrent access to the same key.

```
go run ./cmd/cplt --cardinality 10000 --group 100 --live-ui --duration 10h --rate-limit 5000 curl --concurrency 200 -X 'GET'   'http://127.0.0.1:8008/hello?name=World&locale=ru-RU'   -H 'accept: application/json'
```

Such a command will start a client that will send 10000 different URLs in the loop, trying keep rate of 5000 requests
per second by using up to 200 concurrent requests. Every URLs would be sent in a batch of 100 requests to imitate
concurrency on a single resource.

It will show live performance statistics and overall results.

![cplt screenshot](./resources/screenshots/cplt.png)

Demo app has three modes of operation controlled by `CACHE` environment variable:

* `none` - no caching, all requests are served with involvement of the database,
* `naive` - naive caching with a simple map and TTL of 3 minutes,
* `advanced` - caching using [`github.com/bool64/cache`](https://github.com/bool64/cache) library that implements a
  number of features to improve performance and resiliency, TTL is also 3 minutes.

Application is available at [github.com/vearutop/cache-story](https://github.com/vearutop/cache-story).
If you would like to experiment yourself with it, you can start it with `make start-deps run`.
It depends on `docker-compose` to spin up database, prometheus, grafana (http://localhost:3001) and
jaeger (http://localhost:16686/). You can stop dependencies with `make stop-deps` later.

On my machine I was able to achieve ~500 RPS with no cache. After ~130 concurrent requests DB starts choking
with `Too many connections`. Such result is not great, not terrible, but looks like an improvement opportunity.
Let's see what we can achieve with help of caching.

![Baseline Performance](./resources/screenshots/baseline.png)

With `advanced` cache same laptop was able to show these results.

![Advanced Performance](./resources/screenshots/advanced.png)

```
go run ./cmd/cplt --cardinality 10000 --group 100 --live-ui --duration 10h curl --concurrency 100 -X 'GET'   'http://127.0.0.1:8008/hello?name=World&locale=ru-RU'   -H 'accept: application/json'
```

```
Requests per second: 25064.03
Successful requests: 15692019
Time spent: 10m26.078s

Request latency percentiles:
99%: 28.22ms
95%: 13.87ms
90%: 9.77ms
50%: 2.29ms
```

### Bytes VS Structures

Which one is better?

That depends on the use case, byte cache (or storing data as `[]byte`) have some advantages:

* it grants immutability, because you'll need to decode a new value every time you need it,
* it generally takes less memory, because of less fragmentation,
* it is more friendly to garbage collector, because there is nothing to traverse through,
* it can be easily sent over the wire, because it is exactly what wire expects.

Main disadvantage is the cost of encoding and decoding. In hot loops it can become prohibitively expensive.

Advantages of structures:

* no need to encode/decode a value every time you need it,
* better expressiveness as you can potentially cache things that can not be serialized,

Disadvantages of structure cache:

* mutability, because you reuse same value multiple times it is quite easy to change it without intention,
* memory usage, structures take relatively sparse areas of memory,
* garbage collector pressure, if you have a large set of long-living structures, GC may spend significant time
  traversing them and proving they are still in use.

In this article we will use structure cache.

### Naive Cache

The simplest in-memory cache is a [`map` guarded by a mutex](./internal/infra/cached/naive.go).
When you need a value for a key, you first check if it's in the cache and not expired.
If it's not, you build it from the data source and put it in the cache.
Then you return the value to the caller.

This logic is simple, but it has some drawbacks that may lead to critical issues.

### Concurrent Updates

When multiple callers simultaneously miss the same key, they will all try to build the value. This can lead to a
deadlock or to resource exhaustion [cache stampede](https://en.wikipedia.org/wiki/Cache_stampede) failure.

Additionally, there will be extra latency for all the callers that would try to build the value.
If some of those builds fail, parent callers will fail even though there might be a valid value in the cache.

![Naive Cache Diagram](./resources/screenshots/naive-cache.png)

The issue can be simulated by using low cardinality with high grouping, so that many similar requests are sent at once.

```
go run ./cmd/cplt --cardinality 100 --group 1000 --live-ui --duration 10h --rate-limit 5000 curl --concurrency 150 -X 'GET'   'http://127.0.0.1:8008/hello?name=World&locale=ru-RU'   -H 'accept: application/json'
```

![Key Locking](./resources/screenshots/key-lock.png)

This chart shows application started with `naive` cache and then, on the blue marker it was restarted with `advanced`
cache. As you can see key locking can have a significant impact on performance (mind _Incoming Request Latency_) and
resource usage (mind _DB Operation Rate_).

The solution could be to block parallel builds, so that only one build is in progress at a time. But this would
suffer from contention if there are many concurrent callers asking for a variety of keys.

A better solution is to lock the builds per key, so that one of the callers acquires the lock and owns the build, while
all the others wait for the value.

![Locked Cache Diagram](./resources/screenshots/failover-cache.png)

### Background Updates

When cached entry expires it needs a new value, building new value can be slow. If we do it synchronously, we'll slow
down tail latency (99+ percentile). For cache entries that are in high demand it is feasible to start the build in
advance, even before the value is expired. It can also work if we can afford some level of staleness for the expired
value.

In such case we can immediately serve stale/soon-to-be-expired value and start update in background. One caveat here is
that if depend on parent context, the context may be cancelled right after we served stale value (for example when
parent HTTP request was fulfilled). If we use such context to access database, we'll get a `context canceled` error.
Solution for this problem is to "[detach](https://github.com/bool64/cache/blob/v0.2.5/context.go#L66-L85)" the context
from the parent context and ignore parent cancellation.

Another strategy might be to proactively rebuild cached entries that are soon to be expired, without a parent request,
but this may lead to resource exhaustion due to keeping obsolete cache entries that are of no interest to anybody.

### Expiration Sync

Imagine situation that we start a new instance with TTL cache enabled, cache is empty and almost every request leads to
cache miss and value creation. This will spike the load on the data source and store cached entries with very close
expiration time. Once TTL have passed, the majority of cached entries will expire almost simultaneously, this will lead
to a new load spike. Updated values will have close expiration time again and situation will repeat.

This is a common problem for hot cache entries

Eventually cache entries will come out od sync, but this may take a while.

Solution to this problem is to break the sync by adding jitter to the expiration time.
If expiration jitter is 10% (0.1) it means TTL will vary from `0.95 * TTL` to `1.05 * TTL`. Even such a small jitter
will already help to reduce expiration synchronization.

Here is an example, we're pushing load with high cardinality and high concurrency on the service. It will require many
entries to be available in short period of time, enough to form an expiration spike.

```
go run ./cmd/cplt --cardinality 10000 --group 1 --live-ui --duration 10h --rate-limit 5000 curl --concurrency 200 -X 'GET' 'http://127.0.0.1:8008/hello?name=World&locale=ru-RU' -H 'accept: application/json'
```

![Grafana Chart](./resources/screenshots/expiration-sync.png)

The chart starts with `naive` cache that does not do anything to avoid the sync, second marker indicates service restart
with `advanced` cache that has 10% jitter added to the expiration time. Spikes are wider and shorter and fall faster,
overall service stability is better.

### Errors Caching

When value build fails the easiest thing to do is just return that error to the caller and forget about it.
This can lead to severe issues.

For example, your service works well and handles 10K RPS with help of cache, but suddenly cache builds start to fail (be
it because of temporary database overload, network issue or maybe even logical error like failing validation).
At this point all 10K RPS (instead of usual 100 RPS) will hit data source directly because there will be no cache to
serve.

Minor temporary outage would escalate exponentially.

For high load systems it is very important to cache failures with short TTL to avoid cascading failures.

### Failover Mode

Sometimes serving obsolete value is better than returning error. Especially if obsolete value expired recently and there
is still high chance that it is equal to an up-to-date value.

Failover mode helps to improve resiliency at cost of accuracy, which is often a fair tradeoff in distributed systems.

### Cache Transfer

Cache works best when it has relevant data.
When a new instance of application is started, cache is empty.
Populating helpful data takes time, during this time cache efficiency may degrade significantly.

There are a few ways to work around the issue of "cold" cache.

You can warm up the cache by iterating over the data that is assumed to be useful.
For example, you can fetch recent contents of a database table and store them in cache.
This approach is complex and not always effective.
You need to decide what data to use and rebuild cache entries with a piece of bespoke code.
This may put excessive load on the database (or other sources of data).

You can also avoid this issue by using a shared instance of cache, like redis or memcached.
It has another issue, reading data over the network is much slower, than from local memory.
Also, network bandwidth may become a scalability bottleneck.
Wired data needs to be deserialized which adds on latency and resource usage.

The simple solution to this problem is to transfer cache from active instance to the newly started.
Cached data of active instance naturally has high relevance, because it was populated in response to actual user
requests.
Transferring cache does not need to rebuild data and so it won't abuse data sources.

Usually production systems have multiple instances of application running in parallel.
During deployment, these instances are restarted sequentially, so there is always an instance that is active and has
high quality cache.

Go has a built-in binary serialization format `encoding/gob`. It helps to transfer data over the wire with minimal
effort. The limitation is that it is based on reflection and needs data to have exported fields.

Another caveat with cache transfer is that different versions of application may have different data structures that are
not necessarily compatible. To mitigate this issue, you can fingerprint cached structures (using reflection) and abort
transfer in case of
discrepancy.

<details>
<summary>Here is [a sample implementation](https://github.com/bool64/cache/blob/v0.2.5/gob.go#L49-L90).</summary>

```
// RecursiveTypeHash hashes type of value recursively to ensure structural match.
func recursiveTypeHash(t reflect.Type, h hash.Hash64, met map[reflect.Type]bool) {
	for {
		if t.Kind() != reflect.Ptr {
			break
		}

		t = t.Elem()
	}

	if met[t] {
		return
	}

	met[t] = true

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Skip unexported field.
			if f.Name != "" && (f.Name[0:1] == strings.ToLower(f.Name[0:1])) {
				continue
			}

			if !f.Anonymous {
				_, _ = h.Write([]byte(f.Name))
			}

			recursiveTypeHash(f.Type, h, met)
		}

	case reflect.Slice, reflect.Array:
		recursiveTypeHash(t.Elem(), h, met)
	case reflect.Map:
		recursiveTypeHash(t.Key(), h, met)
		recursiveTypeHash(t.Elem(), h, met)
	default:
		_, _ = h.Write([]byte(t.String()))
	}
}
```

</details>

Transfer can be done with HTTP or any other suitable protocol. In this example, we will use HTTP, served
at [`/debug/transfer-cache`](https://pkg.go.dev/github.com/bool64/cache#HTTPTransfer.Export). Please be aware that cache
may contain sensitive information and should not have exposure to public.

For sake of this example, we can perform transfer with help of a separate instance of an application serving on a
different port.

```
CACHE_TRANSFER_URL=http://127.0.0.1:8008/debug/transfer-cache HTTP_LISTEN_ADDR=127.0.0.1:8009 go run main.go
```

```
2022-05-09T02:33:42.871+0200    INFO    cache/http.go:282       cache restored  {"processed": 10000, "elapsed": "12.963942ms", "speed": "39.564084 MB/s", "bytes": 537846}
2022-05-09T02:33:42.874+0200    INFO    brick/http.go:66        starting server, Swagger UI at http://127.0.0.1:8009/docs
2022-05-09T02:34:01.162+0200    INFO    cache/http.go:175       cache dump finished     {"processed": 10000, "elapsed": "12.654621ms", "bytes": 537846, "speed": "40.530944 MB/s", "name": "greetings", "trace.id": "31aeeb8e9e622b3cd3e1aa29fa3334af", "transaction.id": "a0e8d90542325ab4"}
```

![Cache Transfer](./resources/screenshots/cache-transfer.png)

This chart shows application restarts at blue markers, last two are made with cache transfer. You can see that
performance remains unaffected, while when there is no cache transfer there is a significant warmup penalty.
