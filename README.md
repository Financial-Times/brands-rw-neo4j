# Brands Reader/Writer for Neo4j (brands-rw-neo4j)

__An API for reading/writing brands into Neo4j. Expects the brands json supplied to be in the format that comes out of the brands transformer.__

## Installation

For the first time:

`go get github.com/Financial-Times/brands-rw-neo4j`

or update:

`go get -u github.com/Financial-Times/brands-rw-neo4j`

## Running

`$GOPATH/bin/brands-rw-neo4j --neo-url={neo4jUrl} --port={port} --batchSize=50 --graphiteTCPAddress=graphite.ft.com:2003 --graphitePrefix=content.{env}.brands.rw.neo4j.{hostname} --logMetrics=false

All arguments are optional, they default to a local Neo4j install on the default port (7474), application running on port 8080, batchSize of 1024, graphiteTCPAddress of "" (meaning metrics won't be written to Graphite), graphitePrefix of "" and logMetrics false.

NB: the default batchSize is much higher than the throughput the instance data ingester currently can cope with.

## Updating the model
Use gojson against a transformer endpoint to create a brand struct and update the brand/model.go file. NB: we DO need a separate identifier struct

`curl http://ftaps35629-law1a-eu-t:8080/transformers/brands/ad60f5b2-4306-349d-92d8-cf9d9572a6f6 | gojson -name=brand`

## Building

This service is built and deployed via Jenkins.

<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-build/">Build job</a>
<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-deploy-test/">Deploy to Test job</a>
<a href="http://ftjen10085-lvpr-uk-p:8181/view/JOBS-brands-rw-neo4j/job/brands-rw-neo4j-deploy-prod/">Deploy to Prod job</a>

The build works via git tags. To prepare a new release
- update the version in /puppet/ft-people_rw_neo4j/Modulefile, e.g. to 0.0.12 (If you get a 400 error in your jenkins job you haven't updated this)
- git tag that commit using `git tag 0.0.12`
- `git push --tags`

The deploy also works via git tag and you can also select the environment to deploy to.

## Endpoints
/brands/{uuid}
### PUT
The only mandatory field is the uuid, and the uuid in the body must match the one used on the path.

A successful PUT results in 200.

Queries run in batches. If a batch fails, all failing requests will get a 500 server error response.

Invalid json body input, or uuids that don't match between the path and the body will result in a 400 bad request response.

Example:
`curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" http://localhost:8080/brands/3fa70485-3a57-3b9b-9449-774b001cd965 --data '{"uuid":"3fa70485-3a57-3b9b-9449-774b001cd965", "birthYear": 1974, "salutation": "Mr", "name":"Robert W. Addington", "identifiers":[{ "authority":"http://api.ft.com/system/FACTSET-PPL", "identifierValue":"000BJG-E"}]}'`

### GET
The internal read should return what got written (i.e., this isn't the public brand read API)

If not found, you'll get a 404 response.

Empty fields are omitted from the response.
`curl -H "X-Request-Id: 123" localhost:8080/brands/3fa70485-3a57-3b9b-9449-774b001cd965`

### DELETE
Will return 204 if successful, 404 if not found
`curl -X DELETE -H "X-Request-Id: 123" localhost:8080/brands/3fa70485-3a57-3b9b-9449-774b001cd965`

### Admin endpoints
Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)
Ping: [http://localhost:8080/ping](http://localhost:8080/ping) or [http://localhost:8080/__ping](http://localhost:8080/__ping)

### Loading Data
_*Disclamimer:* this is still a work in progress_
* You will need to install Ruby (2.0+) and the [Nokogiri](http://www.nokogiri.org/) and JSON libraries
  `gem install nokogiri json`
  Be aware that nokogiri uses the libxml2 library and so does some complex compilation, which I've not tried on Windows
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
*The following is historical, see the Loading Data section for current instructions*
There are a couple of sources of data used to extract the current brands and URLs for
* Conversion of trig files to json via jq:
  ```
  rapper -i trig -o json ftdata-brands.trig | jq -c 'to_entries | .[] | { uuid: .key | ltrimstr("http://api.ft.com/things/"), prefLabel: .value["http://www.ft.com/ontology/core/prefLabel"][0].value, parentUUID: .value["http://www.ft.com/ontology/classification/isSubClassificationOf"][0].value | ltrimstr("http://api.ft.com/things/") }'
  ```
* Import converted data to stored
  ```
  cat fromTrig.json | up-restutil put-resources uuid http://localhost:8080/brands
  ```
