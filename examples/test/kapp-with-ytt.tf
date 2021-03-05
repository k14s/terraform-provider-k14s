data "k14s_ytt" "app2" {
  files       = ["ytt-k8s"]
  config_yaml = <<EOF
    #@data/values
    ---
    cm1: "cm1"
    cm2: "cm3"
  EOF
}

resource "k14s_kapp" "app2" {
  app          = "app2"
  namespace    = "default"
  config_yaml  = data.k14s_ytt.app2.result
  diff_changes = true
  debug_logs   = true
}
