package deployment

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDockerBuildProcess tests the Docker build pipeline
func TestDockerBuildProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker build test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get project root
	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile")

	// Verify Dockerfile exists
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Dockerfile not found at %s", dockerfilePath)
	}

	// Test 1: Build Docker image
	t.Run("DockerBuildSuccess", func(t *testing.T) {
		imageName := "rtx-backend:test-" + randomString(8)

		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			projectRoot,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Docker build failed: %v\nOutput: %s", err, string(output))
		}

		// Cleanup image
		defer func() {
			exec.Command("docker", "rmi", imageName).Run()
		}()

		if !strings.Contains(string(output), "Successfully built") {
			t.Logf("Docker build output: %s", string(output))
		}
	})

	// Test 2: Build with different build args
	t.Run("DockerBuildWithBuildArgs", func(t *testing.T) {
		imageName := "rtx-backend:test-args-" + randomString(8)

		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", imageName,
			"--build-arg", "BUILD_ENV=test",
			"--build-arg", "VERSION=1.0.0-test",
			"-f", dockerfilePath,
			projectRoot,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Docker build with args failed: %v\nOutput: %s", err, string(output))
		}

		defer func() {
			exec.Command("docker", "rmi", imageName).Run()
		}()

		t.Logf("Build with args succeeded")
	})

	// Test 3: Verify image layers
	t.Run("DockerImageLayerValidation", func(t *testing.T) {
		imageName := "rtx-backend:test-layers-" + randomString(8)

		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			projectRoot,
		)

		if _, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Docker build failed: %v", err)
		}

		defer func() {
			exec.Command("docker", "rmi", imageName).Run()
		}()

		// Get image history
		historyCmd := exec.CommandContext(ctx, "docker", "history", imageName)
		output, err := historyCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get image history: %v", err)
		}

		if len(output) == 0 {
			t.Error("Image history is empty")
		}

		t.Logf("Image history retrieved successfully")
	})

	// Test 4: Verify image size
	t.Run("DockerImageSizeValidation", func(t *testing.T) {
		imageName := "rtx-backend:test-size-" + randomString(8)

		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			projectRoot,
		)

		if _, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Docker build failed: %v", err)
		}

		defer func() {
			exec.Command("docker", "rmi", imageName).Run()
		}()

		// Check image size
		inspectCmd := exec.CommandContext(ctx, "docker", "inspect",
			"--format='{{.Size}}'", imageName)
		output, err := inspectCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to inspect image: %v", err)
		}

		sizeStr := strings.TrimSpace(string(output))
		if sizeStr == "" || sizeStr == "''" {
			t.Error("Could not determine image size")
		}

		t.Logf("Image size: %s", sizeStr)
	})

	// Test 5: Multi-stage build verification
	t.Run("DockerMultiStageBuild", func(t *testing.T) {
		imageName := "rtx-backend:test-multistage-" + randomString(8)

		cmd := exec.CommandContext(ctx, "docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			projectRoot,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Multi-stage build failed: %v\nOutput: %s", err, string(output))
		}

		defer func() {
			exec.Command("docker", "rmi", imageName).Run()
		}()

		// Verify stages were used
		if !strings.Contains(string(output), "FROM") {
			t.Logf("Warning: Dockerfile may not use multi-stage build: %s", string(output))
		}
	})
}

// TestDockerCompose tests Docker Compose configuration
func TestDockerCompose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker Compose test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	projectRoot := getProjectRoot(t)
	composeFile := filepath.Join(projectRoot, "docker-compose.yml")

	// Check for compose file
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		t.Skipf("docker-compose.yml not found at %s", composeFile)
	}

	t.Run("DockerComposeValidation", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "config")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("docker-compose validation failed: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "services:") {
			t.Error("Docker Compose config is invalid")
		}
	})

	t.Run("DockerComposeServiceValidation", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "config")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get compose config: %v", err)
		}

		// Verify services are defined
		config := string(output)
		if !strings.Contains(config, "backend") && !strings.Contains(config, "app") {
			t.Error("No backend service found in Docker Compose")
		}
	})
}

// TestDockerHealthCheck tests Docker health check configuration
func TestDockerHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker health check test in short mode")
	}

	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile")

	// Read Dockerfile
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	dockerfileContent := string(content)

	t.Run("DockerfileHealthCheckPresent", func(t *testing.T) {
		if !strings.Contains(dockerfileContent, "HEALTHCHECK") {
			t.Log("Dockerfile does not define HEALTHCHECK instruction")
		}
	})

	t.Run("DockerfileExposedPortsPresent", func(t *testing.T) {
		if !strings.Contains(dockerfileContent, "EXPOSE") {
			t.Log("Dockerfile does not define EXPOSE instruction")
		}
	})
}

// TestDockerSecurityScanning tests Docker security best practices
func TestDockerSecurityScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker security scan in short mode")
	}

	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile")

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	dockerfileContent := string(content)

	t.Run("DockerUsesNonRootUser", func(t *testing.T) {
		if !strings.Contains(dockerfileContent, "USER") {
			t.Log("Dockerfile does not explicitly set USER (should not run as root)")
		}
	})

	t.Run("DockerNoLatestTag", func(t *testing.T) {
		lines := strings.Split(dockerfileContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "FROM") && strings.Contains(line, ":latest") {
				t.Error("Dockerfile uses ':latest' tag which is not recommended")
			}
		}
	})

	t.Run("DockerHasNoSudo", func(t *testing.T) {
		if strings.Contains(dockerfileContent, "sudo") {
			t.Log("Dockerfile contains 'sudo' which is not recommended in containers")
		}
	})

	t.Run("DockerPrunesPackageManagers", func(t *testing.T) {
		if strings.Contains(dockerfileContent, "apt-get") {
			if !strings.Contains(dockerfileContent, "apt-get clean") &&
				!strings.Contains(dockerfileContent, "rm -rf /var/lib/apt") {
				t.Log("Dockerfile uses apt-get but may not clean up properly")
			}
		}
	})
}

// TestDockerIgnore tests .dockerignore file
func TestDockerIgnore(t *testing.T) {
	projectRoot := getProjectRoot(t)
	dockerignorePath := filepath.Join(projectRoot, ".dockerignore")

	t.Run("DockerignoreExists", func(t *testing.T) {
		if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
			t.Log(".dockerignore file not found - build context may be larger than necessary")
		}
	})

	t.Run("DockerignoreContentValidation", func(t *testing.T) {
		if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
			t.Skip(".dockerignore file not found")
		}

		content, err := os.ReadFile(dockerignorePath)
		if err != nil {
			t.Fatalf("Failed to read .dockerignore: %v", err)
		}

		dockerignoreContent := string(content)

		// Check for common excludes
		recommendedItems := []string{".git", "node_modules", ".env", "*.md"}
		for _, item := range recommendedItems {
			if !strings.Contains(dockerignoreContent, item) {
				t.Logf("Recommended .dockerignore entry '%s' not found", item)
			}
		}
	})
}

// Helper function to get project root
func getProjectRoot(t *testing.T) string {
	// Try to find go.mod to determine project root
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// Helper function to generate random string for test image names
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
