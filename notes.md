Create a job queue whose workers fetch data from a URL and store the results in a database.
The job queue should expose a REST API for adding jobs and checking their status / results.

Example:

User submits www.google.com to your endpoint.
The user gets back a job id.
Your system fetches www.google.com (the result of which would be HTML) and stores the result.
The user asks for the status of the job id and if the job is complete,
he gets a response that includes the HTML for www.google.com


Finally got dev env set up!
Not quite...

Roadmap/MVP:
1. API that returns a uuid when GET /api/?url="anything" and serves the same data when given any uuid
2. Hook up redis to serve the KV pair
2. Hook up api to rabbitmq
3. GET to /api?url="something" generates a UUID and inserts a job into queue
    maybe also insert into redis uuid with 'processing'
4. Consumer consumes said job then spits out html
5. A bit of documentation

Nice to haves
5. Hook up postgres
6. secondary queue/consumers to feed data into postgres/redis
7. service discovery
8. load balancing
9. docker all the things

Progress Log:
