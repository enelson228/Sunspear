package services

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

type DockerService struct {
	client *client.Client
}

func NewDockerService() (*DockerService, error) {
	// Avoid honoring DOCKER_API_VERSION from env, which can pin an old API version
	// and fail against newer Docker daemons.
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerService{client: cli}, nil
}

func (s *DockerService) Close() error {
	return s.client.Close()
}

// Container operations

func (s *DockerService) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return s.client.ContainerList(ctx, types.ContainerListOptions{All: all})
}

func (s *DockerService) GetContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return s.client.ContainerInspect(ctx, containerID)
}

func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
	return s.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (s *DockerService) StopContainer(ctx context.Context, containerID string, timeout int) error {
	stopTimeout := timeout
	return s.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &stopTimeout})
}

func (s *DockerService) RestartContainer(ctx context.Context, containerID string, timeout int) error {
	stopTimeout := timeout
	return s.client.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &stopTimeout})
}

func (s *DockerService) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return s.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: force})
}

func (s *DockerService) GetContainerLogs(ctx context.Context, containerID string, tail string) (io.ReadCloser, error) {
	return s.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Timestamps: true,
	})
}

func (s *DockerService) GetContainerStats(ctx context.Context, containerID string) (types.ContainerStats, error) {
	return s.client.ContainerStats(ctx, containerID, false)
}

func (s *DockerService) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, containerName string) (container.CreateResponse, error) {
	return s.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
}

func (s *DockerService) RenameContainer(ctx context.Context, containerID string, newName string) error {
	return s.client.ContainerRename(ctx, containerID, newName)
}

// Image operations

func (s *DockerService) ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	return s.client.ImageList(ctx, types.ImageListOptions{All: false})
}

func (s *DockerService) PullImage(ctx context.Context, imageName string) (io.ReadCloser, error) {
	return s.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
}

func (s *DockerService) RemoveImage(ctx context.Context, imageID string, force bool) ([]types.ImageDeleteResponseItem, error) {
	return s.client.ImageRemove(ctx, imageID, types.ImageRemoveOptions{Force: force})
}

func (s *DockerService) SearchImages(ctx context.Context, term string) ([]registry.SearchResult, error) {
	return s.client.ImageSearch(ctx, term, types.ImageSearchOptions{Limit: 25})
}

func (s *DockerService) TagImage(ctx context.Context, imageID string, newRef string) error {
	return s.client.ImageTag(ctx, imageID, newRef)
}

func (s *DockerService) InspectImage(ctx context.Context, imageID string) (types.ImageInspect, error) {
	inspect, _, err := s.client.ImageInspectWithRaw(ctx, imageID)
	return inspect, err
}

func (s *DockerService) GetImageHistory(ctx context.Context, imageID string) ([]types.ImageHistory, error) {
	return s.client.ImageHistory(ctx, imageID)
}

func (s *DockerService) PruneImages(ctx context.Context) (types.ImagesPruneReport, error) {
	return s.client.ImagesPrune(ctx, filters.NewArgs())
}

func (s *DockerService) BuildImage(ctx context.Context, buildContext io.Reader, tags []string) (types.ImageBuildResponse, error) {
	return s.client.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:        tags,
		Remove:      true,
		ForceRemove: true,
	})
}

// System operations

func (s *DockerService) GetSystemInfo(ctx context.Context) (types.Info, error) {
	return s.client.Info(ctx)
}

func (s *DockerService) GetVersion(ctx context.Context) (types.Version, error) {
	return s.client.ServerVersion(ctx)
}

func (s *DockerService) GetDiskUsage(ctx context.Context) (types.DiskUsage, error) {
	return s.client.DiskUsage(ctx, types.DiskUsageOptions{})
}

// Event stream

func (s *DockerService) GetEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	return s.client.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
		),
	})
}
