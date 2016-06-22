# Brands Reader/Writer for Neo4j (brands-rw-neo4j)

__An API for reading/writing brands into Neo4j. Expects the brands json supplied to be in the format that comes out of the brands extractor.__

[Runbook for service](https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/brand-rw-neo4j)

## Developer Notes

### Installation or Update
`go get -u github.com/Financial-Times/brands-rw-neo4j`

### Running
`$GOPATH/bin/brands-rw-neo4j --neo-url={neo4jUrl} --port={port} --batchSize=50 --graphiteTCPAddress=graphite.ft.com:2003 --graphitePrefix=content.{env}.brands.rw.neo4j.{hostname} --logMetrics=false`

All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

### Building
This service is built in CircleCI and deployed via Jenkins.

* The [Jenkins view](http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/) lists the build & deploy jobs
* The [Jenkins job](http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-0-build/) will build whenever a new tag is pushed
* This will get automatically [deployed to test](http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-2-deploy-test/) if the build is successful
* Ut can then be manually pushed to prod via the [deploy to prod](http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-4-deploy-production/) job

## Loading Brand Data
* A google sheet contains the set of brands.
  * [Brand sheet for TEST](https://docs.google.com/spreadsheets/d/1wEdVRLtayZ6-XBfYM3vKAGaOV64cNJD3L8MlLM8-uFY)
  * [Brand sheet for PROD](https://docs.google.com/spreadsheets/d/1Cq8_FyuiSajwn7d9AD0XuJlH1gthxF--5ZFa0JhvNqU)
* This is then published via [Bertha](https://github.com/ft-interactive/bertha/wiki/Tutorial)
* Exposed via the [Brands Transformer](http://git.svc.ft.com:8080/projects/CP/repos/brands-transformer)
* Ingested via the standard [Instance Data Ingester](http://git.svc.ft.com:8080/projects/CP/repos/instance-data-ingester)
* Uses the [WordPress article transformer](http://git.svc.ft.com:8080/projects/CP/repos/wordpress-article-transformer) to clean any html
* Written to NEO4J via *this RW API*
* Finally exposed via the [Public Brands API](https://github.com/Financial-Times/public-brands-api)
* Concordance to TME identifiers is supported by the [Concordance API](https://github.com/Financial-Times/public-concordances-api)

## API Endpoints

This API works, in the main, on the brands/{uuid} path.

### PUT
The only mandatory fields are the uuid, and the alternativeIdentifier uuids (because the uuid is also listed in the alternativeIdentifier uuids list), and the uuid in the body must match the one used on the path. A successful PUT results in 200.
Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

Example:

```
curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" localhost:8080/brands/v --data '{"uuid": "dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54", "prefLabel": "Financial Times","strapline": "Make the right connections", "alternativeIdentifiers":{"uuids": ["dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54","6a2a0170-6afa-4bcc-b427-430268d2ac50"], "TME":["foo","bar"]},"type":"Brand"}'
```

The type field is not currently validated - instead, the Brands Writer writes type Brand and its parent types (Thing, Concept, Classification) as labels for the Brand.

### GET
The internal read should return what got written (i.e., this isn't the public brand read API)

If not found, you'll get a 404 response.

The only field that is omitted if empty is the parentUUID field
```
curl -H "Content-Type: application/json" http://localhost:8080/brands/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54
{"uuid":"dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54","prefLabel":"Financial Times","description":"","strapline":"Make the right connections","descriptionXML":"","_imageUrl":""}
```

### DELETE
Will return 204 if successful, 404 if not found:
```
curl -X DELETE -H "X-Request-Id: 123" localhost:8080/brands/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54
```

### Admin endpoints
* Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)
* Ping: [http://localhost:8080/ping](http://localhost:8080/ping) or [http://localhost:8080/__ping](http://localhost:8080/__ping)
