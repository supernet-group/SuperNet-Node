package docker_utils

import (
	logs "SuperNet-Node/utils/log_utils"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// ImageExists checks if the Docker image with the specified name exists.
func ImageExist(ctx context.Context, cli *client.Client, imageName string) (bool, string) {
	// Retrieve a list of images from Docker daemon
	images, err := cli.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return false, ""
	}
	// Iterate through the list of images retrieved
	for _, image := range images {
		// Check each repo tag of the image
		for _, name := range image.RepoTags {
			// Check if the imageName (e.g., "my-image:latest") contains the current repo tag
			if strings.Contains(imageName, name) {
				logs.Normal(fmt.Sprintf("Image %s exists", imageName))
				return true, image.ID
			}
		}
	}
	logs.Normal(fmt.Sprintf("Image %s does not exist", imageName))
	return false, ""
}

// ContainerExists checks if the Docker container with the specified name exists.
func ContainerExists(ctx context.Context, cli *client.Client, containerName string) (bool, string) {
	// Get the list of all containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, ""
	}
	// Iterate through all containers to find the specified one
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				return true, container.ID
			}
		}
	}
	logs.Normal(fmt.Sprintf("Container %s does not exist", containerName))
	return false, ""
}


// PullImage is a function that pulls a Docker image from a repository.
func PullImage(imageName string) error {
	// Create a command object to execute the docker pull command
	cmd := exec.Command("docker", "pull", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error : Start pulling image: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error : Wait pulling image: %v", err)
	}
	return nil
}

// CreateContainer creates a new container with the specified name and configurations.
// It returns the container ID and an error if the operation fails.
func CreateContainer(ctx context.Context, cli *client.Client, containerName string, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	resp, err := cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// RunContainer runs a Docker container with the specified configuration.
// It ensures the image exists, creates the container, and starts it.
func RunContainer(ctx context.Context, cli *client.Client, containerName string, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	imageName := config.Image
	isCreated, _ := ImageExist(ctx, cli, imageName)
	if !isCreated {
		if err := PullImage(imageName); err != nil {
			return "", err
		}
	}

	containerID, err := CreateContainer(ctx, cli, containerName, config, hostConfig)
	if err != nil {
		return "", err
	}

	logs.Normal(fmt.Sprintf("Start running container %s", containerName))

	if err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return containerID, nil
}

// StopAndRemoveContainer stops and removes a Docker container given its ID.
// It uses the Docker client to issue commands to stop and then remove the container.
func StopAndRemoveContainer(ctx context.Context, cli *client.Client, containerID string) error {
	logs.Normal("Stop and remove container")

	options := container.StopOptions{
		Timeout: nil,
	}
	if err := cli.ContainerStop(ctx, containerID, options); err != nil {
		return err
	}
	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

// DeleteImage deletes an image from the Docker daemon if it exists.
func DeleteImage(ctx context.Context, cli *client.Client, imageName string) error {
	isCreated, imageId := ImageExist(ctx, cli, imageName)
	if isCreated {
		_, err := cli.ImageRemove(ctx, imageId, types.ImageRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDockerImageDirSize retrieves the size of the Docker image directory.
func GetDockerImageDirSize() (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", fmt.Errorf("NewClientWithOpts: %v", err)
	}
	defer cli.Close()

	info, err := cli.Info(context.Background())
	if err != nil {
		return "", fmt.Errorf("cli.Info: %v", err)
	}

	imageDir := fmt.Sprintf("%s/image", info.DockerRootDir)

	size, err := GetDirSize(imageDir)
	if err != nil {
		return "", err
	}

	return size, nil
}

// GetDirSize calculates the size of a directory and returns the size in a human-readable format.
func GetDirSize(dirPath string) (string, error) {
	cmd := exec.Command("du", "-sh", dirPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("du -sh %v: %v", dirPath, err)
	}

	sizeInfo := strings.Fields(string(output))[0]
	return sizeInfo, nil
}
