#!/usr/bin/ruby

require "nokogiri"
require "open-uri"
require 'json'
require 'pp'
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
  puts "loaded #{endpoint['url']}"
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
      puts "Potential rule failure for #{attribute} with #{rule} #{endpoint['uuid']} at #{endpoint['url']}, value will be nil"
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
        data = ProcessEndpoint(endpoint, rule_set)
      rescue RuntimeError => re
        failures[endpoint['uuid']] = "Failure processing #{endpoint['uuid']} at #{endpoint['url']} #{re.message}"
      end
    end
  end
  data['uuid'] = endpoint['uuid']
  endpoint['parentUUID'].nil? ? nil : data['parentUUID'] = endpoint['parentUUID']
  endpoint['prefLabel'].nil? ? nil : data['prefLabel'] = endpoint['prefLabel']
  endpoint['strapline'].nil? ? nil : data['strapline'] = endpoint['strapline']
  endpoint['description'].nil? ? nil : data['description'] = endpoint['description']
  endpoint['descriptionXML'].nil? ? nil : data['descriptionXML'] = endpoint['descriptionXML']
  endpoint['_imageUrl'].nil? ? nil : data['prefLabel'] = endpoint['_imageUrl']
  processed.push(data)
end

puts "Found #{trig_uuids.length()} uuids in trig file and #{brands_uuids.length()} uuids in brands file"

unless(brands_uuids - trig_uuids).empty?
  puts "Found #{(brands_uuids - trig_uuids).length} uuids in trig file but not in brands"
  (brands_uuids - trig_uuids).each do |uuid|
    puts brands_json.endpoints.select {|e| e["uuid"] == uuid }
  end
end
unless (trig_uuids - brands_uuids).empty?
  puts "Found #{(trig_uuids - brands_uuids).length} uuids in trig file but not in brands"
  (trig_uuids - brands_uuids).each do |uuid|
    puts trig_json.select {|trig| trig["uuid"] == uuid }
  end
end


puts "Processed #{processed.length} brands"
File.open("processed.json","wb") do |f|
  json = processed.to_json
  f.write(json)
end
puts "There were #{failures.length} failures"
File.open("failures.json","w") do |f|
  f.write(failures.to_json)
end
