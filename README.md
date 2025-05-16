# Repro repo for testcontainers/testcontainers-go issue #2206

Reproducing https://github.com/testcontainers/testcontainers-go/issues/2206:

1. Environment
    1. Go SDK 1.23+
    1. Docker Engine
1. Run the test
    ```shell
    go test -v -count 1 -timeout 10m ./...
    ```

Output of test when issue happens

```text
=== RUN   TestKafkaContainerStop
    testcontainers_kafka_test.go:31: creating and starting container from image "confluentinc/confluent-local:7.5.0"
2025/05/16 23:29:36 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 28.1.1
  API Version: 1.49
  Operating System: Rocky Linux 9.5 (Blue Onyx)
  Total Memory: 15704 MB
  Testcontainers for Go Version: v0.37.0
  Resolved Docker Host: unix:///var/run/docker.sock
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: ef7cf2e1fa7a887163c2de18319ff64d79d286a09e8b90f3b1221d780f98b987
  Test ProcessID: 007a8a4a-b653-403c-aa98-ca6780e2819f
2025/05/16 23:29:36 🐳 Creating container for image confluentinc/confluent-local:7.5.0
2025/05/16 23:29:36 🐳 Creating container for image testcontainers/ryuk:0.11.0
2025/05/16 23:29:36 ✅ Container created: e5ad55698295
2025/05/16 23:29:36 🐳 Starting container: e5ad55698295
2025/05/16 23:29:36 ✅ Container started: e5ad55698295
2025/05/16 23:29:36 ⏳ Waiting for container id e5ad55698295 image: testcontainers/ryuk:0.11.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms skipInternalCheck:false}
2025/05/16 23:29:36 🔔 Container is ready: e5ad55698295
2025/05/16 23:29:36 ✅ Container created: 797f0626a195
2025/05/16 23:29:36 🐳 Starting container: 797f0626a195
2025/05/16 23:29:41 ✅ Container started: 797f0626a195
2025/05/16 23:29:41 🔔 Container is ready: 797f0626a195
    testcontainers_kafka_test.go:44: successfully started container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc"
    testcontainers_kafka_test.go:46: waiting for 10s before stopping container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc"
    testcontainers_kafka_test.go:59: waiting for container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc" to stop within 5m0s timeout
    testcontainers_kafka_test.go:52: stopping container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc" with 6m0s timeout
2025/05/16 23:29:51 🐳 Stopping container: 797f0626a195
    testcontainers_kafka_test.go:65: container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc" did not stop within 5m0s timeout
    testcontainers_kafka_test.go:76: last 10 log lines of container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc":
        [2025-05-16 20:30:22,580] WARN Using default value http://localhost:8081 for config schema.registry.url. In a future release this config won't have a default value anymore. If you are using Schema Registry, please, specify schema.registry.url explicitly. Requests will fail in a future release if you try to use Schema Registry but have not specified a value for schema.registry.url. An empty value for this property means that the Schema Registry is disabled. (io.confluent.kafkarest.KafkaRestConfig)
        [2025-05-16 20:30:22,587] WARN Using default value http://localhost:8081 for config schema.registry.url. In a future release this config won't have a default value anymore. If you are using Schema Registry, please, specify schema.registry.url explicitly. Requests will fail in a future release if you try to use Schema Registry but have not specified a value for schema.registry.url. An empty value for this property means that the Schema Registry is disabled. (io.confluent.kafkarest.KafkaRestConfig)
        [2025-05-16 20:30:22,598] WARN Using default value http://localhost:8081 for config schema.registry.url. In a future release this config won't have a default value anymore. If you are using Schema Registry, please, specify schema.registry.url explicitly. Requests will fail in a future release if you try to use Schema Registry but have not specified a value for schema.registry.url. An empty value for this property means that the Schema Registry is disabled. (io.confluent.kafkarest.KafkaRestConfig)
        [2025-05-16 20:30:23,090] INFO HV000001: Hibernate Validator 6.1.7.Final (org.hibernate.validator.internal.util.Version)
        [2025-05-16 20:30:23,459] INFO Started o.e.j.s.ServletContextHandler@65e7f52a{/,null,AVAILABLE} (org.eclipse.jetty.server.handler.ContextHandler)
        [2025-05-16 20:30:23,470] INFO Started o.e.j.s.ServletContextHandler@6f815e7f{/ws,null,AVAILABLE} (org.eclipse.jetty.server.handler.ContextHandler)
        [2025-05-16 20:30:23,488] INFO Started NetworkTrafficServerConnector@24ba9639{HTTP/1.1, (http/1.1, h2c)}{0.0.0.0:8082} (org.eclipse.jetty.server.AbstractConnector)
        [2025-05-16 20:30:23,489] INFO Started @2666ms (org.eclipse.jetty.server.Server)
        [2025-05-16 20:30:23,489] INFO Server started, listening for requests... (io.confluent.kafkarest.KafkaRestMain)
    testcontainers_kafka_test.go:38: terminating container "797f0626a195f521c24641ef393833b90832bed4a7c59a6f51c3f144a18d5afc"
2025/05/16 23:34:51 🐳 Stopping container: 797f0626a195
2025/05/16 23:35:01 ✅ Container stopped: 797f0626a195
2025/05/16 23:35:01 ✅ Container stopped: 797f0626a195
2025/05/16 23:35:01 🐳 Terminating container: 797f0626a195
2025/05/16 23:35:01 🚫 Container terminated: 797f0626a195
--- FAIL: TestKafkaContainerStop (325.24s)
FAIL
FAIL	github.com/mabrarov/testcontainers-go-kafka-2748	325.269s
FAIL
```
