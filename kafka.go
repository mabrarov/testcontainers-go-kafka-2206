package testcontainers_go_kafka_2670

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	localControllerPort  = 9094
	localPublicPort      = 9093
	localInternalPort    = 9095
	localLocalhostPort   = 9096
	starterScript        = "/usr/sbin/testcontainers_start.sh"
	starterScriptContent = `#!/bin/bash
# For bitnami/kafka image
export KAFKA_CFG_ADVERTISED_LISTENERS='PLAINTEXT_PUBLIC://%[6]s:%[8]d,PLAINTEXT_INTERNAL://%[7]s:%[4]d,PLAINTEXT_LOCALHOST://localhost:%[5]d'
export KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP='CONTROLLER:PLAINTEXT,PLAINTEXT_PUBLIC:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT,PLAINTEXT_LOCALHOST:PLAINTEXT'
export KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
export KAFKA_CFG_INTER_BROKER_LISTENER_NAME=PLAINTEXT_LOCALHOST
# For apache/kafka and apache/kafka-native images (https://github.com/apache/kafka/blob/trunk/docker/examples/README.md#single-node)
export KAFKA_ADVERTISED_LISTENERS='PLAINTEXT_PUBLIC://%[6]s:%[8]d,PLAINTEXT_INTERNAL://%[7]s:%[4]d,PLAINTEXT_LOCALHOST://localhost:%[5]d'
export KAFKA_LISTENER_SECURITY_PROTOCOL_MAP='CONTROLLER:PLAINTEXT,PLAINTEXT_PUBLIC:PLAINTEXT,PLAINTEXT_INTERNAL:PLAINTEXT,PLAINTEXT_LOCALHOST:PLAINTEXT'
export KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER
export KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT_LOCALHOST
# For bitnami/kafka, apache/kafka and apache/kafka-native images
export KAFKA_LISTENERS='CONTROLLER://:%[2]d,PLAINTEXT_PUBLIC://:%[3]d,PLAINTEXT_INTERNAL://:%[4]d,PLAINTEXT_LOCALHOST://localhost:%[5]d'
export KAFKA_CONTROLLER_QUORUM_VOTERS='0@localhost:%[2]d'
# Run original container entrypoint and command
exec %[1]s
`
)

// KafkaContainer represents the Kafka container type used in the module
type KafkaContainer struct {
	testcontainers.Container
}

// RunKafka creates an instance of the Kafka container type
func RunKafka(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*KafkaContainer, error) {
	publicPort, err := nat.NewPort("tcp", strconv.Itoa(localPublicPort))
	if err != nil {
		return nil, fmt.Errorf("nat.NewPort: %w", err)
	}

	dockerProvider, err := getDockerProvider(opts...)
	if err != nil {
		return nil, fmt.Errorf("getDockerProvider: %w", err)
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{string(publicPort)},
		Env: map[string]string{
			// For bitnami/kafka image
			"KAFKA_CFG_NODE_ID":       "0",
			"KAFKA_CFG_PROCESS_ROLES": "controller,broker",
			// For apache/kafka and apache/kafka-native images (https://github.com/apache/kafka/blob/trunk/docker/examples/README.md#single-node)
			"KAFKA_NODE_ID":                          "0",
			"KAFKA_PROCESS_ROLES":                    "controller,broker",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
		},
		Entrypoint: []string{"sh"},
		Cmd:        []string{"-c", fmt.Sprintf("while [ ! -f %[1]q ]; do sleep 0.1; done; exec %[1]q", starterScript)},
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			{
				PostStarts: []testcontainers.ContainerHook{
					func(ctx context.Context, c testcontainers.Container) error {
						if err := copyStarterScript(ctx, dockerProvider, c, publicPort); err != nil {
							return fmt.Errorf("copy starter script: %w", err)
						}
						return wait.
							ForLog(".*Transition from STARTING to STARTED.*").
							AsRegexp().
							WithStartupTimeout(5*time.Minute).
							WaitUntilReady(ctx, c)
					},
				},
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *KafkaContainer
	if container != nil {
		c = &KafkaContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func getDockerProvider(opts ...testcontainers.ContainerCustomizer) (*testcontainers.DockerProvider, error) {
	// Use a dummy request to get the provider from options.
	var req testcontainers.GenericContainerRequest
	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	logging := req.Logger
	if logging == nil {
		logging = log.Default()
	}
	genericProvider, err := req.ProviderType.GetProvider(testcontainers.WithLogger(logging))
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	if dockerProvider, ok := genericProvider.(*testcontainers.DockerProvider); ok {
		return dockerProvider, nil
	}

	return nil, fmt.Errorf("unknown provider type: %T", genericProvider)
}

// copyStarterScript copies the starter script into the container.
func copyStarterScript(ctx context.Context, dockerProvider *testcontainers.DockerProvider, c testcontainers.Container, localPort nat.Port) error {
	port, err := waitForMappedPort(ctx, c, localPort, time.Minute, 100*time.Millisecond)
	if err != nil {
		return fmt.Errorf("mapped port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return fmt.Errorf("host: %w", err)
	}

	inspect, err := c.Inspect(ctx)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	hostname := inspect.Config.Hostname

	imageInspect, err := dockerProvider.Client().ImageInspect(ctx, inspect.Image)
	if err != nil {
		return fmt.Errorf("image inspect: %w", err)
	}
	containerCmd := buildContainerCmd(imageInspect.Config)

	scriptContent := fmt.Sprintf(starterScriptContent,
		containerCmd,
		localControllerPort,
		localPublicPort,
		localInternalPort,
		localLocalhostPort,
		host,
		hostname,
		port.Int())

	if err := c.CopyToContainer(ctx, []byte(scriptContent), starterScript, 0o755); err != nil {
		return fmt.Errorf("copy to container: %w", err)
	}

	return nil
}

func buildContainerCmd(config *dockercontainer.Config) string {
	entry := append(config.Entrypoint, config.Cmd...)
	for i, cmd := range entry {
		entry[i] = strconv.Quote(cmd)
	}
	return strings.Join(entry, " ")
}

func waitForMappedPort(ctx context.Context, c testcontainers.Container, localPort nat.Port, timeout, interval time.Duration) (nat.Port, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	mappedPort, err := c.MappedPort(ctx, localPort)
	for i := 1; err != nil; i++ {
		select {
		case <-ctx.Done():
			return mappedPort, fmt.Errorf(
				"mapped port: retries: %d, local port: %s, last err: %w, ctx err: %w", i, localPort, err, ctx.Err())
		case <-time.After(interval):
			mappedPort, err = c.MappedPort(ctx, localPort)
		}
	}
	return mappedPort, nil
}

// Brokers retrieves the broker connection strings from Kafka with only one entry,
// defined by the exposed public port.
func (kc *KafkaContainer) Brokers(ctx context.Context) ([]string, error) {
	publicPort, err := nat.NewPort("tcp", strconv.Itoa(localPublicPort))
	if err != nil {
		return nil, fmt.Errorf("nat.NewPort: %w", err)
	}

	host, err := kc.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := kc.MappedPort(ctx, publicPort)
	if err != nil {
		return nil, err
	}

	return []string{fmt.Sprintf("%s:%d", host, port.Int())}, nil
}
