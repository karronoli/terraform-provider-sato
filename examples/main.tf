terraform {
  required_providers {

    sato = {
      version = "0.2"
      source = "karonori.com/personal/sato"
    }
  }
}

provider "sato" {}

module "sato" {
  source = "./sato"

  hardware_address = "00:19:98:ff:ff:ff"
}

output "sato" {
  value = module.sato
}

/*
// for DHCP
resource "sato_network" "pffffff" {
  hardware_address = "00:19:98:80:ce:8d"
  dhcp = true
  rarp = true
}
*/

// for static IP
resource "sato_network" "pffffff" {
  hardware_address = "0019.98ff.ffff"
  dhcp = false
  rarp = false
  ip_address = "192.168.0.10"
  subnet_mask = "255.255.255.0"
  gateway_address = "192.168.0.1"
}
