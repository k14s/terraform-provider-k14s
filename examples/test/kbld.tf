data "k14s_kbld" "tpl1" {
  config_yaml = <<EOF
    images:
    - image: nginx
    - image: mysql
  EOF
}

output "kbld_result" {
  value = data.k14s_kbld.tpl1.result
}
