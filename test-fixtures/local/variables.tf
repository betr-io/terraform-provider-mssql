variable "operating_system" {
  description = "On which operating system is Docker running?"
  default = "Linux"

  validation {
    condition     = contains(["MacOS", "Windows", "Linux"], var.operating_system)
    error_message = "Value must be MacOS, Windows, or Linux."
  }
}