terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
  backend "kubernetes" {
    namespace     = "terraform-backend"
    secret_suffix = "nano-flow"
    config_path   = "~/.kube/config"
  }
}

locals {
  namespace = "rsb-apps"
  name      = "nano-flow"
}

variable "image_tag" {
  type = string
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_persistent_volume_claim_v1" "pvc" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }
  wait_until_bound = false
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "50M"
      }
    }
    storage_class_name = "local-path"
  }
}

resource "kubernetes_service_v1" "service" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }
  spec {
    selector = {
      app = kubernetes_deployment_v1.deploy.spec[0].selector[0].match_labels.app
    }

    port {
      protocol    = "TCP"
      port        = 50051
      target_port = 50051
    }
  }
}

resource "kubernetes_secret_v1" "env" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }
  data = {
    NANO_DATABASE_PATH = "/data/nano-flow.db"
  }
}

resource "kubernetes_deployment_v1" "deploy" {
  metadata {
    name      = local.name
    namespace = local.namespace
  }

  wait_for_rollout = false

  spec {
    replicas = 1
    selector {
      match_labels = {
        app = local.name
      }
    }

    template {
      metadata {
        labels = {
          app = local.name
        }
      }
      spec {
        container {
          name  = local.name
          image = "maxrsb/nano-flow:${var.image_tag}"
          env_from {
            secret_ref {
              name = kubernetes_secret_v1.env.metadata[0].name
            }
          }
          resources {
            limits = {
              cpu    = "400m"
              memory = "100Mi"
            }
            requests = {
              cpu    = "20m"
              memory = "5Mi"
            }
          }
          volume_mount {
            name       = local.name
            mount_path = "/data"
          }
        }

        volume {
          name = local.name
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim_v1.pvc.metadata[0].name
          }
        }

      }
    }
  }
}

output "nano_flow_url" {
  value = "http://${local.name}.${local.namespace}.svc.cluster.local:${kubernetes_service_v1.service.spec[0].port[0].port}"
}
