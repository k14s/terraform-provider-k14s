resource "k14s_kapp" "app1" {
  app       = "app1"
  namespace = "default"

  config_yaml = <<EOF
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: str-123
    data:
      str: str-124
  EOF

  diff_changes = true
}

resource "k14s_kapp" "app3" {
  app       = "app3"
  namespace = "default"

  config_yaml = <<EOF
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: str-app3
    data:
      str: str-124
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: str-app3-another
    data:
      str: str-124
  EOF

  diff_changes = true

  deploy {
    raw_options = [
      "--wait=false",
    ]
  }
}
