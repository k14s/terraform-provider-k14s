provider "k14s" {
  kapp {
    kubeconfig {
      from_env = true
    }
  }
}
