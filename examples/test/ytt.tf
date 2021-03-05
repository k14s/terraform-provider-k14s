data "k14s_ytt" "example1" {
  files = ["ytt-example"]

  config_yaml = <<EOF
    #@data/values
    ---
    str: tfstr
    #@overlay/match missing_ok=True
    map:
      nested: true
      nested2: true
  EOF

  values = {
    str = "tfstr2"
    # Will be interpreted as string by ytt
    "map.nested" = "true"
  }

  values_yaml = {
    # Will be interpreted as boolean by ytt
    "map.nested2" = "true"
  }
}

// See `terraform output` for the result
output "result" {
  value = data.k14s_ytt.example1.result
}
