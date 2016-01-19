# People Reader/Writer for Neo4j (people-rw-neo4j)

__An API for reading/writing people into Neo4j. Expects the people json supplied to be in the format that comes out of the people transformer.__

## Installation

For the first time:

`go get github.com/Financial-Times/people-rw-neo4j`

or update:

`go get -u github.com/Financial-Times/people-rw-neo4j`

## Running

`$GOPATH/bin/people-rw-neo4j --neo-url={neo4jUrl} --port={port} --batchSize=50 --graphiteTCPAddress=graphite.ft.com:2003 --graphitePrefix=content.{env}.people.rw.neo4j.{hostname} --logMetrics=false

All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

NB: the default batchSize is much higher than the throughput the instance data ingester currently can cope with.

## Updating the model
Use gojson against a transformer endpoint to create a person struct and update the person/model.go file. NB: we DO need a separate identifier struct

`curl http://ftaps35629-law1a-eu-t:8080/transformers/people/ad60f5b2-4306-349d-92d8-cf9d9572a6f6 | gojson -name=person`

## Building

This service is built and deployed via Jenkins.

<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-people-rw-neo4j/job/people-rw-neo4j-build/">Build job</a>
<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-people-rw-neo4j/job/people-rw-neo4j-deploy-test/">Deploy to Test job</a>
<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-people-rw-neo4j/job/people-rw-neo4j-deploy-prod/">Deploy to Prod job</a>

The build works via git tags. To prepare a new release
- update the version in /puppet/ft-people_rw_neo4j/Modulefile, e.g. to 0.0.12 (If you get a 400 error in your jenkins job you haven't updated this)
- git tag that commit using `git tag 0.0.12`
- `git push --tags`

The deploy also works via git tag and you can also select the environment to deploy to.

## Endpoints
/people/{uuid}
### PUT
The only mandatory field is the uuid, and the uuid in the body must match the one used on the path.

Every request results in an attempt to update that person: unlike with GraphDB there is no check on whether the person already exists and whether there are any changes between what's there and what's being written. We just do a MERGE which is Neo4j for create if not there, update if it is there.

A successful PUT results in 200.

We run queries in batches. If a batch fails, all failing requests will get a 500 server error response.

Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

Example:
`curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" localhost:8080/people/3fa70485-3a57-3b9b-9449-774b001cd965 --data '{"uuid":"3fa70485-3a57-3b9b-9449-774b001cd965", "birthYear": 1974, "salutation": "Mr", "name":"Robert W. Addington", "identifiers":[{ "authority":"http://api.ft.com/system/FACTSET-PPL", "identifierValue":"000BJG-E"}]}'`

### GET
Thie internal read should return what got written (i.e., this isn't the public person read API)

If not found, you'll get a 404 response.

Empty fields are omitted from the response.
`curl -H "X-Request-Id: 123" localhost:8080/people/3fa70485-3a57-3b9b-9449-774b001cd965`

### DELETE
Will return 204 if successful, 404 if not found
`curl -XDELETE -H "X-Request-Id: 123" localhost:8080/people/3fa70485-3a57-3b9b-9449-774b001cd965`

### Admin endpoints
Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)

Ping: [http://localhost:8080/ping](http://localhost:8080/ping) or [http://localhost:8080/__ping](http://localhost:8080/__ping)
