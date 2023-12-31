package main

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Message struct pour stocker le message de réponse JSON
type FlinkSessionJob struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StartTime string `json:"startTime"`
}

// Get operator tenants list
func ListFlinkJobs() ([]FlinkSessionJob, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)

	resourceId := schema.GroupVersionResource{
		Group:    "flink.apache.org",
		Version:  "v1beta1",
		Resource: "flinksessionjobs",
	}

	list, err := dynamic.Resource(resourceId).Namespace("flink-operator").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	jobs := []FlinkSessionJob{}
	for _, item := range list.Items {
		name, _, err := unstructured.NestedString(item.UnstructuredContent(), "metadata", "name")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		status, _, err := unstructured.NestedString(item.UnstructuredContent(), "status", "lifecycleState")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		jobId, _, err := unstructured.NestedString(item.UnstructuredContent(), "status", "jobStatus", "jobId")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		startTime, _, err := unstructured.NestedString(item.UnstructuredContent(), "status", "jobStatus", "startTime")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		jobs = append(jobs, FlinkSessionJob{Name: name, Status: status, ID: jobId, StartTime: startTime})
	}

	return jobs, nil
}

func UpdateFlinkSessionJob(resourceName, statusValue string) error {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)

	resourceId := schema.GroupVersionResource{
		Group:    "flink.apache.org",
		Version:  "v1beta1",
		Resource: "flinksessionjobs",
	}

	// Fetch the existing object
	resource, err := dynamic.Resource(resourceId).Namespace("flink-operator").Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting resource: %v\n", err)
		return err
	}

	// Modify the status field
	resource.Object["spec"].(map[string]interface{})["job"].(map[string]interface{})["state"] = statusValue

	// Update the resource
	_, err = dynamic.Resource(resourceId).Namespace("flink-operator").Update(ctx, resource, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error updating resource status: %v\n", err)
		return err
	}

	return nil
}
