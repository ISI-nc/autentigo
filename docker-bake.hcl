variable "GOPROXY" {
  default = null
}

variable "GONOSUMDB" {
  default = null
}

group "default" {
  targets = ["apps"]
}

target "_common" {
  dockerfile = "./Dockerfile"
  args = {
    GOPROXY = GOPROXY
    GONOSUMDB = GONOSUMDB
  }
  platforms = ["linux/amd64", "linux/arm64"]
}

target "apps" {
  inherits = ["_common"]
  target = "apps"
}
