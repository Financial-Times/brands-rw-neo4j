# Brands Reader/Writer for Neo4j (brands-rw-neo4j)

__An API for reading/writing brands into Neo4j. Expects the brands json supplied to be in the format that comes out of the brands extractor.__

[Runbook for service](https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/brand-rw-neo4j)

## Installation or Update
`go get -u github.com/Financial-Times/brands-rw-neo4j`

## Running
`$GOPATH/bin/brands-rw-neo4j --neo-url={neo4jUrl} --port={port} --batchSize=50 --graphiteTCPAddress=graphite.ft.com:2003 --graphitePrefix=content.{env}.brands.rw.neo4j.{hostname} --logMetrics=false`

All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

## Building
*TO BE COMPLETED / DOCUMENTED*

This service is built in CircleCI and deployed via Jenkins.

## Endpoints

/brands/{uuid}

### PUT
The only mandatory field is the uuid, and the uuid in the body must match the one used on the path. A successful PUT results in 200.
Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

Example:
```
curl -H "Content-Type: application/json" -X PUT http://localhost:8080/brands/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54 --data '{"uuid": "dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54","prefLabel": "Financial Times","strapline": "Make the right connections"}'
```

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

### Loading Data
_*Disclaimer:* this is still a work in progress (the version of WPAT is not in prod)_
* You will need to install Ruby (2.0+) and the [Nokogiri](http://www.nokogiri.org/)
  `gem install nokogiri`
  Be aware that nokogiri uses the libxml2 library and so does some complex compilation, which can be problematic.
* You will also need a version of the WordPress Article Transformer that has the new `HtmlTransformerResource` endpoint, which is currently available in the a [HtmlTransformer branch](http://git.svc.ft.com/projects/CP/repos/wordpress-article-transformer/commits/b1d23060f717364b40a6506f74429f9a290a2b71)
* Run the wordpress-article-transformer, passing in the following command line arguments: `server wordpress-article-transformer.yaml`
  By default this will start the transformer on port `14080`
* `cd` into the `extractor` subdirectory and launch the `BrandExtractor.rb` script.
  `ruby BrandExtractor.rb`
* Assuming all goes well a `processed.json` file is written, any major failures are written to a `failures.json` file
* Use the `up-restutil` and [jq](https://stedolan.github.io/jq/) tool to load the data via the write API
  `jq '.[]' -c processed.json | up-restutil put-resources uuid http://localhost:8080/brands`
* You can then use the private read API, public read API, or Neo4J browser to look at the data.


### Notes
_The following is historical. The resulting file is currently stored as [extractor/fromTrig.json](https://github.com/Financial-Times/brands-rw-neo4j/blob/master/extractor/fromTrig.json)_
* Get the currently known set of Brands from [Semantic Data Ontologies Repository](http://git.svc.ft.com/projects/CP/repos/semantic-data-ontologies/browse/src/main/resources/ontology/ft/instance_data/ftdata-brands.trig)
* Conversion of trig files to json via [rapper](https://apps.ubuntu.com/cat/applications/precise/raptor2-utils/) and [jq](https://stedolan.github.io/jq/):
  ```
  rapper -i trig -o json ftdata-brands.trig | jq -c 'to_entries | .[] | { uuid: .key | ltrimstr("http://api.ft.com/things/"), prefLabel: .value["http://www.ft.com/ontology/core/prefLabel"][0].value, parentUUID: .value["http://www.ft.com/ontology/classification/isSubClassificationOf"][0].value | ltrimstr("http://api.ft.com/things/") }'
  ```
