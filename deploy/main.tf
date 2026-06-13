terraform {
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = ">= 4.0.0"
    }
  }
}

# 1. Authentication Layer
provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

# 2. Virtual Cloud Network (VCN)
resource "oci_core_vcn" "bot_vcn" {
  cidr_block     = "10.0.0.0/16"
  compartment_id = var.compartment_ocid
  display_name   = "smieci-sms-network"
  dns_label      = "smiecisms"
}

# 3. Internet Gateway
resource "oci_core_internet_gateway" "bot_ig" {
  compartment_id = var.compartment_ocid
  display_name   = "smieci-sms-gateway"
  vcn_id         = oci_core_vcn.bot_vcn.id
}

# 4. Route Table
resource "oci_core_default_route_table" "bot_rt" {
  manage_default_resource_id = oci_core_vcn.bot_vcn.default_route_table_id
  display_name               = "smieci-sms-route-table"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.bot_ig.id
  }
}

# 5. Security List (Firewall Rules)
resource "oci_core_default_security_list" "bot_security" {
  manage_default_resource_id = oci_core_vcn.bot_vcn.default_security_list_id
  display_name               = "smieci-sms-security-list"

  # Outbound Rule (Allow your app to hit external APIs like Telegram)
  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  # Inbound Rule: SSH Access
  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 22
      max = 22
    }
  }

  # Inbound Rule: HTTP (Required for Let's Encrypt certificate challenge)
  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 80
      max = 80
    }
  }

  # Inbound Rule: HTTPS (Secure Telegram Webhooks)
  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 443
      max = 443
    }
  }
}

# 6. Subnet Configuration
resource "oci_core_subnet" "bot_subnet" {
  cidr_block        = "10.0.1.0/24"
  compartment_id    = var.compartment_ocid
  vcn_id            = oci_core_vcn.bot_vcn.id
  display_name      = "smieci-sms-public-subnet"
  dns_label         = "botsubnet"
  route_table_id    = oci_core_vcn.bot_vcn.default_route_table_id
  security_list_ids = [oci_core_vcn.bot_vcn.default_security_list_id]
}

resource "oci_core_instance" "free_arm_vm" {
  availability_domain = var.availability_domain
  compartment_id      = var.compartment_ocid
  display_name        = "smieci-sms-production-server"
  
  shape = "VM.Standard.E2.1.Micro"
#   shape_config {
#     ocpus         = 1  # Always Free max
#     memory_in_gbs = 1 # Always Free max
#   }

  source_details {
    source_type             = "image"
    source_id               = var.image_ocid # Use an Ubuntu ARM minimal image OCID for your region
    boot_volume_size_in_gbs = 50             # Anything up to 200GB is free
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.bot_subnet.id
    assign_public_ip = true
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key_path)
    
    # Cloud-init script to automatically configure host OS on boot
    user_data = base64encode(<<-EOF
      #!/bin/bash
      # 1. Direct system clock to Warsaw Rules
      timedatectl set-timezone Europe/Warsaw

      # 2. Open Ubuntu OS-level firewall for web traffic (Crucial for Oracle Linux/Ubuntu images)
      iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
      iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
      netfilter-persistent save

      # 3. Update indices and install stable Docker engine tooling + curl
      apt-get update
      apt-get install -y docker.io docker-compose curl

      # 4. Create deployment directories and configure folder owners
      mkdir -p /app
      chown -R ubuntu:ubuntu /app

      # 5. Bind the default SSH user directly to the Docker engine interface
      usermod -aG docker ubuntu


    EOF
    )
  }
}

output "instance_public_ip" {
  value       = oci_core_instance.free_arm_vm.public_ip
  description = "Connect here to deploy code or register your webhook link"
}