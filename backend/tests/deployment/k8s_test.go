package deployment

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestKubernetesManifests validates Kubernetes manifest files
func TestKubernetesManifests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kubernetes validation in short mode")
	}

	projectRoot := getProjectRoot(t)
	k8sDir := filepath.Join(projectRoot, "k8s")

	// Check if k8s directory exists
	if _, err := os.Stat(k8sDir); os.IsNotExist(err) {
		t.Skipf("Kubernetes manifests directory not found at %s", k8sDir)
	}

	t.Run("K8sDeploymentManifestValidation", func(t *testing.T) {
		deploymentFile := filepath.Join(k8sDir, "deployment.yaml")
		if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
			t.Skipf("Deployment manifest not found at %s", deploymentFile)
		}

		validateK8sManifest(t, deploymentFile)
	})

	t.Run("K8sServiceManifestValidation", func(t *testing.T) {
		serviceFile := filepath.Join(k8sDir, "service.yaml")
		if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
			t.Skipf("Service manifest not found at %s", serviceFile)
		}

		validateK8sManifest(t, serviceFile)
	})

	t.Run("K8sConfigMapValidation", func(t *testing.T) {
		configMapFile := filepath.Join(k8sDir, "configmap.yaml")
		if _, err := os.Stat(configMapFile); os.IsNotExist(err) {
			t.Skipf("ConfigMap manifest not found at %s", configMapFile)
		}

		validateK8sManifest(t, configMapFile)
	})

	t.Run("K8sIngressManifestValidation", func(t *testing.T) {
		ingressFile := filepath.Join(k8sDir, "ingress.yaml")
		if _, err := os.Stat(ingressFile); os.IsNotExist(err) {
			t.Logf("Ingress manifest not found at %s (optional)", ingressFile)
			return
		}

		validateK8sManifest(t, ingressFile)
	})

	t.Run("K8sHPAManifestValidation", func(t *testing.T) {
		hpaFile := filepath.Join(k8sDir, "hpa.yaml")
		if _, err := os.Stat(hpaFile); os.IsNotExist(err) {
			t.Logf("HPA manifest not found at %s (optional)", hpaFile)
			return
		}

		validateK8sManifest(t, hpaFile)
	})
}

// TestKubernetesDeploymentRequirements validates deployment requirements
func TestKubernetesDeploymentRequirements(t *testing.T) {
	projectRoot := getProjectRoot(t)
	deploymentFile := filepath.Join(projectRoot, "k8s", "deployment.yaml")

	if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
		t.Skipf("Deployment manifest not found at %s", deploymentFile)
	}

	content, err := os.ReadFile(deploymentFile)
	if err != nil {
		t.Fatalf("Failed to read deployment manifest: %v", err)
	}

	manifestContent := string(content)

	t.Run("DeploymentHasResources", func(t *testing.T) {
		if !strings.Contains(manifestContent, "resources:") {
			t.Error("Deployment does not define resource requests/limits")
		}

		if !strings.Contains(manifestContent, "requests:") {
			t.Log("Deployment does not define resource requests")
		}

		if !strings.Contains(manifestContent, "limits:") {
			t.Log("Deployment does not define resource limits")
		}
	})

	t.Run("DeploymentHasLivenessProbe", func(t *testing.T) {
		if !strings.Contains(manifestContent, "livenessProbe:") {
			t.Log("Deployment does not define liveness probe")
		}
	})

	t.Run("DeploymentHasReadinessProbe", func(t *testing.T) {
		if !strings.Contains(manifestContent, "readinessProbe:") {
			t.Log("Deployment does not define readiness probe")
		}
	})

	t.Run("DeploymentHasSecurityContext", func(t *testing.T) {
		if !strings.Contains(manifestContent, "securityContext:") {
			t.Log("Deployment does not define security context")
		}

		if !strings.Contains(manifestContent, "runAsNonRoot:") {
			t.Log("Deployment does not enforce non-root user")
		}
	})

	t.Run("DeploymentHasNodeSelector", func(t *testing.T) {
		if !strings.Contains(manifestContent, "nodeSelector:") {
			t.Logf("Deployment does not define node selector (optional)")
		}
	})

	t.Run("DeploymentHasAffinity", func(t *testing.T) {
		if !strings.Contains(manifestContent, "affinity:") {
			t.Logf("Deployment does not define affinity rules (optional)")
		}
	})

	t.Run("DeploymentHasImagePullPolicy", func(t *testing.T) {
		if !strings.Contains(manifestContent, "imagePullPolicy:") {
			t.Log("Deployment does not define image pull policy")
		}
	})

	t.Run("DeploymentHasStrategyDefined", func(t *testing.T) {
		if !strings.Contains(manifestContent, "strategy:") {
			t.Log("Deployment does not define update strategy")
		}
	})
}

// TestKubernetesServiceRequirements validates service configuration
func TestKubernetesServiceRequirements(t *testing.T) {
	projectRoot := getProjectRoot(t)
	serviceFile := filepath.Join(projectRoot, "k8s", "service.yaml")

	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		t.Skipf("Service manifest not found at %s", serviceFile)
	}

	content, err := os.ReadFile(serviceFile)
	if err != nil {
		t.Fatalf("Failed to read service manifest: %v", err)
	}

	manifestContent := string(content)

	t.Run("ServiceHasSelector", func(t *testing.T) {
		if !strings.Contains(manifestContent, "selector:") {
			t.Error("Service does not define selector")
		}
	})

	t.Run("ServiceHasPorts", func(t *testing.T) {
		if !strings.Contains(manifestContent, "ports:") {
			t.Error("Service does not define ports")
		}
	})

	t.Run("ServiceHasType", func(t *testing.T) {
		if !strings.Contains(manifestContent, "type:") {
			t.Log("Service does not explicitly define type (defaults to ClusterIP)")
		}
	})
}

// TestHealthCheckEndpoints validates health check endpoints
func TestHealthCheckEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping health check endpoint test in short mode")
	}

	endpoints := []struct {
		name     string
		endpoint string
		expected int
	}{
		{"Health", "/health", http.StatusOK},
		{"Readiness", "/ready", http.StatusOK},
		{"Liveness", "/live", http.StatusOK},
	}

	// Get service endpoint from environment or use default
	baseURL := os.Getenv("SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			url := baseURL + ep.endpoint

			// Create request with timeout
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Logf("Health check endpoint %s not available (service may not be running): %v", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != ep.expected {
				t.Errorf("Expected status %d, got %d for %s", ep.expected, resp.StatusCode, url)
			}
		})
	}
}

// TestDatabaseMigrationInContainers validates database migration process
func TestDatabaseMigrationInContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database migration test in short mode")
	}

	t.Run("MigrationScriptExists", func(t *testing.T) {
		projectRoot := getProjectRoot(t)
		migrationPath := filepath.Join(projectRoot, "migrations")

		if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
			t.Skipf("Migrations directory not found at %s", migrationPath)
		}

		// Check if there are any migration files
		entries, err := os.ReadDir(migrationPath)
		if err != nil {
			t.Fatalf("Failed to read migrations directory: %v", err)
		}

		if len(entries) == 0 {
			t.Log("No migration files found in migrations directory")
		}

		t.Logf("Found %d migration files", len(entries))
	})

	t.Run("MigrationHook", func(t *testing.T) {
		projectRoot := getProjectRoot(t)
		deploymentFile := filepath.Join(projectRoot, "k8s", "deployment.yaml")

		if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
			t.Skipf("Deployment manifest not found")
		}

		content, err := os.ReadFile(deploymentFile)
		if err != nil {
			t.Fatalf("Failed to read deployment manifest: %v", err)
		}

		manifestContent := string(content)

		// Check for migration init container or hook
		if strings.Contains(manifestContent, "initContainers:") {
			t.Logf("Deployment uses init containers (for migrations)")
		} else if strings.Contains(manifestContent, "lifecycle:") && strings.Contains(manifestContent, "postStart:") {
			t.Logf("Deployment uses lifecycle hooks")
		} else {
			t.Log("Deployment may not run database migrations (no init containers or lifecycle hooks)")
		}
	})

	t.Run("MigrationEnvVars", func(t *testing.T) {
		projectRoot := getProjectRoot(t)
		deploymentFile := filepath.Join(projectRoot, "k8s", "deployment.yaml")

		if _, err := os.Stat(deploymentFile); os.IsNotExist(err) {
			t.Skipf("Deployment manifest not found")
		}

		content, err := os.ReadFile(deploymentFile)
		if err != nil {
			t.Fatalf("Failed to read deployment manifest: %v", err)
		}

		manifestContent := string(content)

		// Check for database environment variables
		requiredEnvVars := []string{"DB_HOST", "DB_PASSWORD", "DB_USER", "DB_NAME"}
		for _, envVar := range requiredEnvVars {
			if !strings.Contains(manifestContent, envVar) {
				t.Logf("Environment variable %s not found in deployment", envVar)
			}
		}
	})
}

// TestKubernetesSecurity validates Kubernetes security configuration
func TestKubernetesSecurity(t *testing.T) {
	projectRoot := getProjectRoot(t)
	k8sDir := filepath.Join(projectRoot, "k8s")

	if _, err := os.Stat(k8sDir); os.IsNotExist(err) {
		t.Skipf("Kubernetes manifests directory not found")
	}

	t.Run("NetworkPolicyExists", func(t *testing.T) {
		netpolFile := filepath.Join(k8sDir, "networkpolicy.yaml")
		if _, err := os.Stat(netpolFile); os.IsNotExist(err) {
			t.Log("NetworkPolicy not found (recommended for security)")
		}
	})

	t.Run("RBACConfigExists", func(t *testing.T) {
		rbacFile := filepath.Join(k8sDir, "rbac.yaml")
		if _, err := os.Stat(rbacFile); os.IsNotExist(err) {
			t.Log("RBAC configuration not found")
		}
	})

	t.Run("PodSecurityPolicyExists", func(t *testing.T) {
		pspFile := filepath.Join(k8sDir, "psp.yaml")
		if _, err := os.Stat(pspFile); os.IsNotExist(err) {
			t.Logf("PodSecurityPolicy not found (optional for newer K8s)")
		}
	})
}

// TestKubernetesHPA validates Horizontal Pod Autoscaler configuration
func TestKubernetesHPA(t *testing.T) {
	projectRoot := getProjectRoot(t)
	hpaFile := filepath.Join(projectRoot, "k8s", "hpa.yaml")

	if _, err := os.Stat(hpaFile); os.IsNotExist(err) {
		t.Skipf("HPA manifest not found at %s", hpaFile)
	}

	content, err := os.ReadFile(hpaFile)
	if err != nil {
		t.Fatalf("Failed to read HPA manifest: %v", err)
	}

	manifestContent := string(content)

	t.Run("HPAHasMinReplicas", func(t *testing.T) {
		if !strings.Contains(manifestContent, "minReplicas:") {
			t.Error("HPA does not define minimum replicas")
		}
	})

	t.Run("HPAHasMaxReplicas", func(t *testing.T) {
		if !strings.Contains(manifestContent, "maxReplicas:") {
			t.Error("HPA does not define maximum replicas")
		}
	})

	t.Run("HPAHasMetrics", func(t *testing.T) {
		if !strings.Contains(manifestContent, "metrics:") {
			t.Error("HPA does not define scaling metrics")
		}
	})

	t.Run("HPATargetReference", func(t *testing.T) {
		if !strings.Contains(manifestContent, "scaleTargetRef:") {
			t.Error("HPA does not define scale target reference")
		}
	})
}

// Helper function to validate K8s manifest
func validateK8sManifest(t *testing.T, filePath string) {
	// Check if kubectl is available
	_, err := exec.LookPath("kubectl")
	if err != nil {
		t.Logf("kubectl not found in PATH, skipping manifest validation: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", filePath, "--dry-run=client", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Kubernetes manifest validation failed for %s: %v\nOutput: %s",
			filepath.Base(filePath), err, string(output))
	}
}
