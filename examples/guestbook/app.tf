provider "k14s" {
  kapp {
    kubeconfig {
      from_env = true
    }
  }
}

data "k14s_ytt" "guestbook" {
  files                   = ["config"]
  ignore_unknown_comments = true

  # Configure all deployments to have 1 replica
  config_yaml = <<EOF
    #@ load("@ytt:overlay", "overlay")
    #@overlay/match by=overlay.subset({"kind":"Deployment"}),expects="1+"
    ---
    spec:
      replicas: 1
  EOF
}

resource "k14s_kapp" "guestbook" {
  app          = "guestbook"
  namespace    = "default"
  config_yaml  = data.k14s_ytt.guestbook.result
  diff_changes = true
}
