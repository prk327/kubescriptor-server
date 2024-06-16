// backend/main.go
package main

import (
	"bytes"
	"context"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ScriptRequest struct {
	Script string `json:"script"`
}

type ScriptResponse struct {
	Output string `json:"output"`
}

func main() {
	// Initialize the Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Initialize Gin
	r := gin.Default()

	r.POST("/api/run-script", func(c *gin.Context) {
		var request ScriptRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create a Kubernetes Job to run the script
		job := &v1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bash-script-job",
			},
			Spec: v1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "bash-script-container",
								Image:   "alpine",
								Command: []string{"/bin/sh", "-c", request.Script},
							},
						},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
		}

		jobClient := clientset.BatchV1().Jobs("default")
		_, err := jobClient.Create(context.Background(), job, metav1.CreateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Wait for the job to complete and get the logs
		// This is a simplified example and doesn't handle job status checking or log retrieval
		time.Sleep(10 * time.Second) // Wait for job to complete
		cmd := exec.Command("kubectl", "logs", "job/bash-script-job")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := ScriptResponse{Output: out.String()}
		c.JSON(http.StatusOK, response)
	})

	r.Run() // Listen and serve on 0.0.0.0:8080
}
