#WARNING:
This is not at all production ready.
There are MANY problems with this due to time constraints.
I'll attempt to address a few concerns below, but would be happy to talk in depth about what I would have liked to do given time.

#NOTE:
I was having issues with my environment (I just bought a new mac and had to set up my dev environment)
I tried to label places where code reuse would have normally occurred but I didn't want to waste time debugging my environment.

#MVP_DESIGN
A simple rabbitmq consumer and API service with two endpoints, GET /api/:uuid and POST /api?query
1. POST --> Inserts a worker into rabbit on the 'job' queue, returning a uuid, sets the value in the cache to 'processing'
`{ id: <uuid>, query: <url> }`
2. Worker subscribes to 'job' queue, reads messages, fetches the html and inserts it into Redis
3. GET fetches the currently set value in redis

#Setup
1. Install docker, golang
2. Run setup.sh to launch rabbit/redis as containers
3. Set configs and run the API/Fetcher

```bash
    . ./env.sh
    cd api
    go run main.go
```

```bash
    # in a separate tab
    . ./env.sh
    cd fetcher
    go run main.go
```

#API
POST /api?query='http://url.com' --> uuid of worker
GET /api/:uuid --> html || status of the worker

#TODO:
Theres so much room for improvement.
1. If redis goes down, we lose everything --> Use postgres as the source of truth
2. The queries are assumed to be in a specific format (ie. with the scheme provided) but no validation is being done
3. We try to process the message but since there's no validation there's not a good way to retry on failures.
    With validation we can be more confident about requeueing messages so data is not lost
4. Depending on how read/write heavy this service gets, we can and should split read/write services out. (would be happy to elaborate in detail)
5. Config, Secrets, Service discovery
6. Streamlining the dev/deployment process
7. Docker all the things!