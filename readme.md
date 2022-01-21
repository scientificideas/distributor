Distributor is a general deterministic work distribution service. It allows you to define what each service should do.

<br>

#### Architecture

[Russian](https://gitlab.n-t.io/atmz/distributor/-/blob/master/architecture_ru.md)
| [English](https://gitlab.n-t.io/atmz/distributor/-/blob/master/architecture_en.md)

<br>

#### Configuration

| CLI argument           | ENV variable           | description                                                    | example                            | default value      |
|------------------------|------------------------|----------------------------------------------------------------|------------------------------------|--------------------|
| connection             | CONNECTION             | Fabric connection profile                                      | -connection=connection.yaml        | connection.yaml    |
| user                   | USER                   | Fabric user                                                    | -user=User1                        | User1              |
| org                    | ORG                    | Fabric org                                                     | -org=atomyze                       | atomyze            |
| redis-addrs            | REDIS_ADDRS            | Redis nodes addresses                                          | -redis-addrs="redis-6379:6379,redis-6380:6380,redis-6381:6381,redis-6382:6382,redis-6383:6383,redis-6384:6384" | 0.0.0.0:6379       |
| redis-pass             | REDIS_PASS             | Redis password                                                 | -redis-pass="secret"               | ""                 |
| redis-tls              | REDIS_TLS              | enable TLS for communication with Redis                        | -redis-tls=true                    | false              |
| redis-rootca-certs     | REDIS_ROOTCA_CERTS     | comma-separated root CA's certificates list for TLS with Redis | -redis-rootca-certs=/path/to/ca1.pem,/path/to/ca2.pem | ""                 |                   
| poll-interval          | POLL_INTERVAL          | services ping interval                                         | -poll-interval=500ms               | 1000ms             |
| ping-timeout           | PING_TIMEOUT           | ping request timeout                                           | -ping-timeout=500ms                | 1000ms             |
| distribution-namespace | DISTRIBUTION_NAMESPACE | key in storage where work distribution data is stored          | -distribution-namespace=sys-matching-table | sys-matching-table |
| services-namespace     | SERVICES_NAMESPACE     | key in storage where services list is stored                   | -services-namespace=sys-robots-list | sys-robots-list    |
| workunits-namespace    | WORKUNITS_NAMESPACE    | key in storage where work units list is stored                 | -workunits-namespace=sys-channels  | sys-channels       |
| prom-port              | PROM_PORT              | Prometheus metrics port                                        | -prom-port=8473                    | 9090               |
| ka-time                | KA_TIME                | KeepAlive time                                                 | -ka-time=10s                       | 10s                |
| ka-timeout             | KA_TIMEOUT             | KeepAlive timeout                                              | -ka-timeout=20s                    | 20s                |
| ka-permit-without-stream | KA_PERMIT_WITHOUT_STREAM | KeepAlive param: if true, client sends keepalive pings even with no active RPCs; if false, when there are no active RPCs, Time and Timeout will be ignored and no keepalive pings will be sent   | -ka-permit-without-stream=false    | false              |
| config-type            | -                      | which type of config to use                                    | -config-type=args or -config-type=env | args               |

<br>

#### Startup example

Build:

    docker build -t distributor .

Run:

     docker run --restart unless-stopped --name distributor --network sample_net \ 
     -e POLL_INTERVAL=200ms -e LOG=debug -e PING_TIMEOUT=300ms -e -DISTRIBUTION_NAMESPACE=matching-table-namespace \
     -e SERVICES_NAMESPACES=oneservices-list-namespace,anotherservices-list-namespace -e WORKUNITS_NAMESPACE=workunits-namespace \
     -e REDIS_PASS="" -e REDIS_ADDRS=redis-6379:6379,redis-6380:6380,redis-6381:6381,redis-6382:6382,redis-6383:6383,redis-6384:6384 \ 
     -d distributor
    
