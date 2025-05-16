package testcontainers_go_kafka_2670

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

var (
	imageFlag       = flag.String("image", "confluentinc/confluent-local:7.5.0", "image with Kafka")
	stopTimeoutFlag = flag.Duration("timeout", 5*time.Minute, "Kafka stop timeout")
	tailFlag        = flag.Int("tail", 10, "number of last log lines of Kafka container to output")
)

func TestKafkaContainerStop(t *testing.T) {
	const pauseBeforeStop = 10 * time.Second

	image := *imageFlag
	tailLength := *tailFlag
	stopTimeout := *stopTimeoutFlag

	ctx := context.Background()
	t.Logf("creating and starting container from image %q", image)
	container, err := kafka.Run(ctx, image)
	if err != nil {
		t.Fatalf("container start failed: %s", err)
	}

	defer func() {
		t.Logf("terminating container %q", container.GetContainerID())
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("container %q termination failed: %v", container.GetContainerID(), err)
		}
	}()

	t.Logf("successfully started container %q", container.GetContainerID())

	t.Logf("waiting for %v before stopping container %q", pauseBeforeStop, container.GetContainerID())
	time.Sleep(pauseBeforeStop)

	stopped := make(chan struct{})
	go func() {
		timeout := stopTimeout + time.Minute
		t.Logf("stopping container %q with %v timeout", container.GetContainerID(), timeout)
		if err := container.Stop(ctx, &timeout); err != nil {
			t.Errorf("container %q stop failed: %v", container.GetContainerID(), err)
		}
		stopped <- struct{}{}
	}()

	t.Logf("waiting for container %q to stop within %v timeout", container.GetContainerID(), stopTimeout)

	select {
	case <-stopped:
		t.Logf("successfully stopped container %q within %v timeout", container.GetContainerID(), stopTimeout)
	case <-time.After(stopTimeout):
		t.Errorf("container %q did not stop within %v timeout", container.GetContainerID(), stopTimeout)
	}

	printContainerLogTail(t, ctx, container, tailLength)
}

func printContainerLogTail(t *testing.T, ctx context.Context, container testcontainers.Container, num int) {
	lines, err := containerLogTail(ctx, container, num)
	if err != nil {
		t.Fatalf("failed to get container %q log line(s): %v", container.GetContainerID(), err)
	}
	t.Logf("last %d log lines of container %q:\n%s", num, container.GetContainerID(), strings.Join(lines, "\n"))
}

func containerLogTail(ctx context.Context, container testcontainers.Container, num int) ([]string, error) {
	logs, err := container.Logs(ctx)
	if err != nil {
		return nil, fmt.Errorf("logs of container %q: %v", container.GetContainerID(), err)
	}
	defer func() {
		_ = logs.Close()
	}()
	return lastLines(logs, num)
}

func lastLines(reader io.Reader, num int) ([]string, error) {
	if num <= 0 {
		return nil, nil
	}
	lines := make([]string, 0, num)
	last := num - 1
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) < num {
			lines = append(lines, line)
		} else {
			copy(lines[0:last], lines[1:num])
			lines[last] = line
		}
	}
	return lines, scanner.Err()
}
