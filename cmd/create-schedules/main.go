package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"pli-agent-api/workflows"

	"go.temporal.io/sdk/client"
)

func main() {
	// Parse command-line flags
	temporalHost := flag.String("host", "localhost:7233", "Temporal server address")
	temporalNamespace := flag.String("namespace", "default", "Temporal namespace")
	scheduleType := flag.String("type", "all", "Schedule type to create (all, license-deactivation)")
	flag.Parse()

	// Connect to Temporal
	c, err := client.Dial(client.Options{
		HostPort:  *temporalHost,
		Namespace: *temporalNamespace,
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	defer c.Close()

	log.Printf("Connected to Temporal at %s (namespace: %s)", *temporalHost, *temporalNamespace)

	switch *scheduleType {
	case "all":
		createLicenseDeactivationSchedule(c)
	case "license-deactivation":
		createLicenseDeactivationSchedule(c)
	default:
		log.Fatalf("Unknown schedule type: %s", *scheduleType)
	}

	log.Println("All schedules created successfully!")
}

// createLicenseDeactivationSchedule creates the daily license expiry check schedule
// WF-AGT-PRF-007: License Deactivation Workflow
// BR-AGT-PRF-013: Auto-Deactivation on License Expiry
func createLicenseDeactivationSchedule(c client.Client) {
	scheduleID := "daily-license-expiry-check"

	log.Printf("Creating schedule: %s", scheduleID)

	// Check if schedule already exists
	handle := c.ScheduleClient().GetHandle(context.Background(), scheduleID)
	_, err := handle.Describe(context.Background())
	if err == nil {
		log.Printf("Schedule %s already exists. Deleting and recreating...", scheduleID)
		err = handle.Delete(context.Background())
		if err != nil {
			log.Printf("Warning: Failed to delete existing schedule: %v", err)
		}
	}

	// Create schedule for daily license expiry check
	// Every day at 2:00 AM IST (20:30 UTC previous day)
	schedule := client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			// Cron: Every day at 2:00 AM IST (20:30 UTC)
			// Adjust timezone as needed
			CronExpressions: []string{"30 20 * * *"}, // 20:30 UTC = 02:00 IST
			// Alternative: Use calendar spec for more complex schedules
			// Calendars: []client.ScheduleCalendarSpec{
			//     {
			//         Hour:   []client.ScheduleRange{{Start: 2}},
			//         Minute: []client.ScheduleRange{{Start: 0}},
			//     },
			// },
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "license-deactivation-workflow",
			Workflow:  workflows.LicenseDeactivationWorkflow,
			TaskQueue: "agent-profile-task-queue",
			Args: []interface{}{
				workflows.DeactivationInput{
					BatchDate: time.Now(),
					DryRun:    false,
				},
			},
			// Workflow options
			WorkflowExecutionTimeout: 30 * time.Minute,
			WorkflowRunTimeout:       25 * time.Minute,
			WorkflowTaskTimeout:      10 * time.Second,
		},
		// Overlap policy: Skip if previous run is still executing
		Overlap: client.ScheduleOverlapPolicySkip,
		// Catchup window: Don't run missed schedules
		CatchupWindow: 0,
		// Pause on failure: Don't pause after failures
		PauseOnFailure: false,
	}

	handle, err = c.ScheduleClient().Create(context.Background(), schedule)
	if err != nil {
		log.Fatalf("Failed to create schedule %s: %v", scheduleID, err)
	}

	// Describe the created schedule
	desc, err := handle.Describe(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to describe schedule: %v", err)
	} else {
		log.Printf("✓ Schedule created successfully: %s", scheduleID)
		log.Printf("  - Cron: %v", desc.Schedule.Spec.CronExpressions)
		log.Printf("  - Task Queue: %s", schedule.Action.(*client.ScheduleWorkflowAction).TaskQueue)
		log.Printf("  - Overlap Policy: %s", schedule.Overlap)
		log.Printf("  - Next run: %v", desc.Info.NextActionTimes)
	}
}

// Future schedules can be added here:
//
// func createMonthlyReportSchedule(c client.Client) {
//     // Monthly report generation
// }
//
// func createWeeklyReconciliationSchedule(c client.Client) {
//     // Weekly data reconciliation
// }
