package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// We need a struct to hold the data coming into the workflow
type RideRequest struct {
	RideID string
	Lat    float64
	Lng    float64
}

func RideLifecycleWorkflow(ctx workflow.Context, req RideRequest) (string, error) {
	// 1. Set up Activity Options (Timeouts and Retries)
	// We want our activities to timeout after 10 seconds, and retry up to 3 times if they fail.
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumAttempts:    3,
	}
	
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	
	// Apply these options to our context
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// We need to tell Temporal what our activities are named, so it can find them
	var activities *RideActivities 

	// ---------------------------------------------------------
	
	// 1. Call the EstimatePriceActivity
	var price int
	err := workflow.ExecuteActivity(ctx, activities.EstimatePriceActivity, req.RideID).Get(ctx, &price)
	if err != nil {
		return "", err
	}

	// 2. Call the FindAndAssignDriverActivity
	var driverID string
	err = workflow.ExecuteActivity(ctx, activities.FindAndAssignDriverActivity, req.RideID).Get(ctx, &driverID)
	if err != nil {
		return "", err
	}

	// 3. Call the UpdateTripStatusActivity to mark it as "ASSIGNED"
	err = workflow.ExecuteActivity(ctx, activities.UpdateTripStatusActivity, req.RideID, "ASSIGNED").Get(ctx, nil)
	if err != nil {
		return "", err
	}

	// If we got this far, the workflow succeeded!
	return driverID, nil
}
