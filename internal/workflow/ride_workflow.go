package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type RideRequest struct {
	RideID string
	Lat    float64
	Lng    float64
}

func RideLifecycleWorkflow(ctx workflow.Context, req RideRequest) (string, error) {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumAttempts:    3,
	}
	
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	
	ctx = workflow.WithActivityOptions(ctx, activityOpts)
	var activities *RideActivities 

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

	// NEW: Wait for the Driver to accept the ride!
	signalChan := workflow.GetSignalChannel(ctx, "DriverAcceptedSignal")
	timerFuture := workflow.NewTimer(ctx, 30*time.Second)
	selector := workflow.NewSelector(ctx)

	var signalReceived bool
	
	// If the signal arrives first, this block runs
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil) // We drain the channel
		signalReceived = true
	})

	// If the timer expires first, this block runs
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		// Do nothing, signalReceived remains false
	})

	// This is where the workflow actually PAUSES and sleeps!
	selector.Select(ctx)

	if !signalReceived {
		return "", fmt.Errorf("timeout waiting for driver to accept")
	}


	// 3. Call the UpdateTripStatusActivity to mark it as "ASSIGNED"
	err = workflow.ExecuteActivity(ctx, activities.UpdateTripStatusActivity, req.RideID, "ASSIGNED").Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return driverID, nil
}
