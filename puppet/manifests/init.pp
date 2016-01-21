class brands-rw-neo4j {

  $configParameters = hiera('configParameters','')
  
  class { "go_service_profile" :
    service_name => "brands-rw-neo4j"
    configParameters => "$configParameters"
  }

}
