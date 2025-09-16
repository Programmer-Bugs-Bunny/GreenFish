data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "atlas_loader.go",
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "postgres://admin:123456@192.168.1.5:5432/migration?sslmode=disable"
  url = "postgres://admin:123456@192.168.1.5:5432/dev_db?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "production" {
  src = data.external_schema.gorm.url
  url = "${var.DATABASE_URL}"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

variable "DATABASE_URL" {
  type    = string
  default = ""
}