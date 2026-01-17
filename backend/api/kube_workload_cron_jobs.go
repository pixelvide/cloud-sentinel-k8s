package api

import (
	"fmt"
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetCronJobs lists cronjobs for a given namespace and context
func GetCronJobs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// Parse namespaces
	namespaces := ParseNamespaces(ns)

	type CronJobInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Schedule     string `json:"schedule"`
		TimeZone     string `json:"timezone,omitempty"`
		Suspend      bool   `json:"suspend"`
		Active       int    `json:"active"`
		LastSchedule string `json:"last_schedule"`
		Age          string `json:"age"`
	}

	var cronjobs []CronJobInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		list, err := clientset.BatchV1().CronJobs(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, cj := range list.Items {
			suspend := false
			if cj.Spec.Suspend != nil {
				suspend = *cj.Spec.Suspend
			}
			lastSchedule := ""
			if cj.Status.LastScheduleTime != nil {
				lastSchedule = cj.Status.LastScheduleTime.Time.Format(time.RFC3339)
			}
			timezone := ""
			if cj.Spec.TimeZone != nil {
				timezone = *cj.Spec.TimeZone
			}
			cronjobs = append(cronjobs, CronJobInfo{
				Name:         cj.Name,
				Namespace:    cj.Namespace,
				Schedule:     cj.Spec.Schedule,
				TimeZone:     timezone,
				Suspend:      suspend,
				Active:       len(cj.Status.Active),
				LastSchedule: lastSchedule,
				Age:          cj.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"cronjobs": cronjobs})
}

// ToggleCronJobSuspend toggles the suspend state of a CronJob
func ToggleCronJobSuspend(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	namespace := c.Query("namespace")
	name := c.Query("name")

	var input struct {
		Suspend bool `json:"suspend"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	patch := []byte(fmt.Sprintf(`{"spec":{"suspend":%t}}`, input.Suspend))
	_, err = clientset.BatchV1().CronJobs(namespace).Patch(c.Request.Context(), name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch cronjob: " + err.Error()})
		return
	}

	// Record audit log
	action := "Suspend CronJob"
	if !input.Suspend {
		action = "Resume CronJob"
	}
	RecordAuditLog(c, action, gin.H{
		"context":   ctxName,
		"namespace": namespace,
		"name":      name,
		"suspend":   input.Suspend,
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("CronJob %s updated", name)})
}

// TriggerCronJob creates a Job from a CronJob
func TriggerCronJob(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	namespace := c.Query("namespace")
	name := c.Query("name")

	var input struct {
		JobName string `json:"jobName"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.JobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job name is required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// 1. Get the CronJob to extract the template
	cronJob, err := clientset.BatchV1().CronJobs(namespace).Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cronjob: " + err.Error()})
		return
	}

	// 2. Create the Job object
	// References: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/cronjob/utils.go#L38
	annotations := make(map[string]string)
	annotations["cronjob.kubernetes.io/instantiate"] = "manual"
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		annotations[k] = v
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        input.JobName,
			Namespace:   namespace,
			Annotations: annotations,
			Labels:      cronJob.Spec.JobTemplate.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cronJob.Name,
					UID:        cronJob.UID,
				},
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	// 3. Create the Job
	result, err := clientset.BatchV1().Jobs(namespace).Create(c.Request.Context(), job, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job: " + err.Error()})
		return
	}

	// 4. Audit Log
	RecordAuditLog(c, "Trigger CronJob", gin.H{
		"context":     ctxName,
		"namespace":   namespace,
		"cronJobName": name,
		"jobName":     input.JobName,
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Job %s created", result.Name), "jobName": result.Name})
}
