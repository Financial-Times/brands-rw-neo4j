class brands-rw-neo4j {

  class { "go_service_profile" :
    service_name => "brands-rw-neo4j"
    configParameters => hiera('configParameters','')
  }
  
}
