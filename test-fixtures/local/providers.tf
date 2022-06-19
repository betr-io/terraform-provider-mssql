provider "docker" {
  host = var.operating_system == "Windows" ? "npipe:////.//pipe//docker_engine" : "unix:///var/run/docker.sock"
}

provider "local" {}
