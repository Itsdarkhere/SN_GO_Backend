job "bitclout-backend" {
  datacenters = ["fin-yx"]
  type = "service"

  update {
    stagger = "90s"
    max_parallel = 1
  }

  group "bclt-backend" {
    
    restart {
      attempts = 3
      interval = "5m"
      delay = "60s"
      mode = "delay"
    }

    volume "social" {
      type = "host"
      read_only = false
      source = "social"
    }

    task "bclt-core" {
      driver = "docker"
      config {
        image = "registry.gitlab.com/love4src/backend:[[ .commit_sha ]]"
        command = "run"
        auth {
          password = "7J4iX1XvJ9BUGZbQ3Uz2",
          username = "gitlab+deploy-token-466088"
        }
      }

      env {
          GLOG_V=0
          TESTNET=false
          EXTERNAL_IPS="65.21.92.186:${NOMAD_PORT_protocol}"
          MINER_PUBLIC_KEYS="BC1YLiVHMzQDAaXHy6Q1vb8Yf8WDRGSWExuELW9bYpfGuUahGym6Pwf"
          ADMIN_PUBLIC_KEYS="BC1YLhwpmWkgk2iM9yTSxzgUVhYjgessSPTiVHkkK9pMrhweqJnWrvK"
          SUPER_ADMIN_KEYS="BC1YLiVHMzQDAaXHy6Q1vb8Yf8WDRGSWExuELW9bYpfGuUahGym6Pwf"
          NUM_MINING_THREADS=1
          PROTOCOL_PORT=${NOMAD_PORT_protocol}
          API_PORT="${NOMAD_PORT_json}"
          RATE_LIMIT_FEERATE=400
          MIN_FEERATE=100000
          TARGET_OUTBOUND_PEERS=64
          MAX_PEERS=125
          DATA_DIR="/db"
          ONE_INBOUND_PER_IP=true
          STALL_TIMEOUT_SECONDS=900
          STARTER_BITCLOUT_NANOS=0
          SECURE_HEADER_DEVELOPMENT=true
          ACCESS_CONTROL_ALLOW_ORIGINS="*"
          MIN_SATOSHIS_FOR_PROFILE=100000
          MAX_BLOCK_TEMPLATES_CACHE=0
          MIN_BLOCK_UPDATE_INTERVAL=10
          IGNORE_INBOUND_INVS=false
          IGNORE_UNMINED_BITCOIN=false
          PRIVATE_MODE=false
          READ_ONLY_MODE=false
          TXINDEX=true
          DISABLE_NETWORKING=false
          SUPPORT_EMAIL="andrzej@yexperiment.com"
          BLOCK_CYPHER_API_KEY="{{with secret "l4s-kv/config"}}{{.Data.data.block_cypher_key}}{{end}}"
          BUY_BITCLOUT_SEED="{{with secret "l4s-kv/buy_clout"}}{{.Data.data.source_clout}}{{end}}"
          BUY_BITCLOUT_BTC_ADDRESS="{{with secret "l4s-kv/buy_clout"}}{{.Data.data.target_btc}}{{end}}"
          LOG_DB_SUMMARY_SNAPSHOTS=false
      }

      volume_mount {
        volume = "social"
        destination = "/db"
        read_only = false
      }

      resources {
        cpu    = 16384
        memory = 16384
        network {
          mbits = 10
          port "protocol" {
            static = 17000
          }
          port "json" { }
        }
      }

      service {
        name = "bitclout-backend"
        port = "protocol"

        tags = [
          "internal-proxy.enable=true",

          "internal-proxy.http.routers.bitclout-core.rule=Host(`api.love4src.com`) || (Host(`love4src.com`) && PathPref>          "internal-proxy.http.routers.bitclout-core.tls=true",
          "internal-proxy.http.routers.bitclout-core.tls.certresolver=astroresolver"
        ]

        check {
          name     = "BitClout Core JSON API"
          type     = "http"
          path     = "/"
          interval = "120s"
          timeout  = "30s"
          check_restart {
            limit = 3
            grace = "30s"
            ignore_warnings = true
          }
        }
      }

      service {
        name = "bitclout-backend"
        port = "json"
        
        tags = [
          "internal-proxy.enable=true",

          "internal-proxy.http.routers.bitclout-core.rule=Host(`api.love4src.com`) || (Host(`love4src.com`) && PathPrefix(`/api/`) )",
          "internal-proxy.http.routers.bitclout-core.tls=true",
          "internal-proxy.http.routers.bitclout-core.tls.certresolver=astroresolver"
        ]

        check {
          name     = "BitClout Core JSON API"
          type     = "http"
          path     = "/"
          interval = "120s"
          timeout  = "30s"
          check_restart {
            limit = 3
            grace = "30s"
            ignore_warnings = true
          }
        }
      }
    }
  }
}
