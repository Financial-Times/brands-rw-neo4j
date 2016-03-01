#!/usr/bin/ruby

require "nokogiri"
require "open-uri"
require 'json'
require 'logger'
require 'net/http'

# Open the fromTrig.json file this gives lists of brands known according to grpahdb
# Open the BrandsExtractor.json - this is the list of brands we will actually scrape fromTrig
# Compare the two and put out a warning message

# For each endpoint in BrandsExtractor.json run the ruleSet
# If the ruleset doesn't exist then warn / fail

# The tricky one is the bodyXML conversion that needs to happen.
# Might be able to call the Java WordPress transformer if we run this as jruby
# See https://kgilmersden.wordpress.com/2010/09/30/call-your-java-method-from-your-ruby-class/
# If not then need to copy (subset of) business logic from transformer (aaaggh)

# WARNING - HBD

@log = Logger.new(STDOUT)

trig_file = File.read('fromTrig.json')
brand_file = File.read('BrandExtractor.json')

trig_json = JSON.parse(trig_file)
brands_json = JSON.parse(brand_file)

trig_uuids = Array.new
brands_uuids = Array.new
trig_json.each { |x| trig_uuids.push(x['uuid']) }
brands_json['endpoints'].each { |x| brands_uuids.push(x['uuid'])}


# This is the main function, most of the other code is just about looping and printing messages...
def ProcessEndpoint(endpoint, rule_set)
  result = {}
  doc = Nokogiri::HTML(open(endpoint['url']))
  rule_set['rules'].each do |attribute, rule|
    thing = doc.at_css(rule['select'])
    unless thing.nil?
      value = nil
      case rule['extract']
      when "inner_html"
        value = thing.inner_html
      when "text"
        value = thing.text
      when "style"
        value = thing.attribute('style')
      end
      if rule['filter']
        filtered = ""
        Regexp.new(rule['filter']).match(value) {|md| md.captures.each {|c| filtered += c}}
        value = filtered
      end
      if rule['transformer']
        value = Transform(value, rule['transformer'])
      end
    else
      @log.warn "Nil value for #{attribute} with #{rule}"
      result[attribute] = nil
    end
    result[attribute]=value
  end
  return result
end

def Transform(text, transformer)
  uri = URI.parse("http://localhost:14080/content-transformer")
  http = Net::HTTP.new(uri.host,uri.port)
  req = Net::HTTP::Post.new(uri.path)
  req.body = text
  req['Content-Type'] = "text/html"
  res = http.request(req)
  #puts res.body.force_encoding('iso-8859-1')
  return res.body.force_encoding('iso-8859-1')
end

failures = {}
processed = []

brands_json['endpoints'].each do |endpoint|
  data={}
  unless endpoint['ruleSet'] == nil
    rule_set = brands_json['ruleSets'][endpoint['ruleSet']]
    if rule_set.nil?
      failures[endpoint['uuid']] = {message: "Unable to find a ruleset to process brand #{endpoint['uuid']} at #{endpoint['url']}"}
    else
      begin
        @log.info "Processing #{endpoint['uuid']} at #{endpoint['url']}"
        data = ProcessEndpoint(endpoint, rule_set)
      rescue RuntimeError => re
        failures[endpoint['uuid']] = "Failure processing #{endpoint['uuid']} at #{endpoint['url']} #{re.message}"
      end
    end
  end
  data['uuid'] = endpoint['uuid']
  data['parentUUID'] = endpoint['parentUUID'] || data['parentUUID']
  data['prefLabel'] = endpoint['prefLabel'] || data['prefLabel']
  data['strapline'] = endpoint['strapline'] || data['strapline']
  data['description'] = endpoint['description'] || data['description']
  data['descriptionXML'] = endpoint['descriptionXML'] || data['descriptionXML']
  data['prefLabel'] = endpoint['_imageUrl'] || data['prefLabel']
  processed.push(data)
end

@log.info "Found #{trig_uuids.length()} uuids in trig file and #{brands_uuids.length()} uuids in brands file"

unless(brands_uuids - trig_uuids).empty?
  @log.info "Found #{(brands_uuids - trig_uuids).length} uuids in trig file but not in brands"
  (brands_uuids - trig_uuids).each do |uuid|
    puts brands_json.endpoints.select {|e| e["uuid"] == uuid }
  end
end
unless (trig_uuids - brands_uuids).empty?
  @log.info "Found #{(trig_uuids - brands_uuids).length} uuids in trig file but not in brands"
  (trig_uuids - brands_uuids).each do |uuid|
    puts trig_json.select {|trig| trig["uuid"] == uuid }
  end
end

File.open("processed.json","w") do |f|
  json = processed.to_json
  f.write(json)
end
File.open("failures.json","w") do |f|
  f.write(failures.to_json)
end

@log.error "There were #{failures.length} failures, see failures.json for details" if failures.length >0
@log.info "Total of #{processed.length} brands processed and written to processed.json"
