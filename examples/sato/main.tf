terraform {
  required_providers {
    sato = {
      version = "0.2"
      source  = "karonori.com/personal/sato"
    }
  }
}

variable "hardware_address" {
  type    = string
  default = "00-19-98-ff-ff-ff"
}

data "sato_networks" "all" {}

output "all_networks" {
  value = data.sato_networks.all.networks
}

output "network" {
  value = {
    for network in data.sato_networks.all.networks :
    network.hardware_address => network
    if network.hardware_address == var.hardware_address
  }
}
