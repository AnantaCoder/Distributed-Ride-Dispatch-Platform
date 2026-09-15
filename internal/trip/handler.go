package trip

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
	tripv1 "github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/gen/trip/v1"
)

// Handler implements the generated ConnectRPC TripServiceHandler interface.
// It translates network requests (protobuf) into calls to your internal Service logic.
type Handler struct {
	service Service
}

// NewHandler creates a new ConnectRPC handler for the Trip service.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetTrip handles the GetTrip RPC.
func (h *Handler) GetTrip(ctx context.Context, req *connect.Request[tripv1.GetTripRequest]) (*connect.Response[tripv1.GetTripResponse], error) {

	
	TripID , err := uuid.Parse(req.Msg.TripId)

	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument,err)
	}

	trip , err := h.service.GetTrip(ctx, TripID)

	if err != nil {
		return nil  , connect.NewError(connect.CodeInvalidArgument,err)
	}

	driverIdStr := ""
	if trip.DriverID != nil {
		driverIdStr = trip.DriverID.String()
	}

	response := &tripv1.GetTripResponse{
		Trip: &tripv1.Trip{
			Id:       trip.ID.String(),
			UserId:   trip.PassengerID.String(),
			DriverId: driverIdStr,
			PickupLocation: &tripv1.Location{
				Latitude:  trip.PickupLat,
				Longitude: trip.PickupLng,
			},
			DropoffLocation: &tripv1.Location{
				Latitude:  trip.DropoffLat,
				Longitude: trip.DropoffLng,
			},
			Status: tripv1.TripStatus(tripv1.TripStatus_value[trip.Status]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: trip.CreatedAt.Unix(),
			},
			UpdatedAt: &timestamppb.Timestamp{
				Seconds: trip.UpdatedAt.Unix(),
			},
		},
	}
	

	
	return connect.NewResponse(response), nil
}

// CreateTrip handles the CreateTrip RPC.
func (h *Handler) CreateTrip(ctx context.Context, req *connect.Request[tripv1.CreateTripRequest]) (*connect.Response[tripv1.CreateTripResponse], error) {
	

	userId , err := uuid.Parse(req.Msg.UserId)
	
	if err!= nil{
		return nil , connect.NewError(connect.CodeInvalidArgument,err)
	}

	trip, err := h.service.RequestRide(
		ctx,
		userId,
		req.Msg.PickupLocation.Latitude,
		req.Msg.PickupLocation.Longitude,
		req.Msg.DropoffLocation.Latitude,
		req.Msg.DropoffLocation.Longitude,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	driverIdStr := ""
	if trip.DriverID != nil {
		driverIdStr = trip.DriverID.String()
	}

	response := &tripv1.CreateTripResponse{
		Trip: &tripv1.Trip{
			Id:       trip.ID.String(),
			UserId:   trip.PassengerID.String(),
			DriverId: driverIdStr,
			PickupLocation: &tripv1.Location{
				Latitude:  trip.PickupLat,
				Longitude: trip.PickupLng,
			},
			DropoffLocation: &tripv1.Location{
				Latitude:  trip.DropoffLat,
				Longitude: trip.DropoffLng,
			},
			Status: tripv1.TripStatus(tripv1.TripStatus_value[trip.Status]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: trip.CreatedAt.Unix(),
			},
			UpdatedAt: &timestamppb.Timestamp{
				Seconds: trip.UpdatedAt.Unix(),
			},
		},
	}
	
	return connect.NewResponse(response), nil
}

// UpdateTrip handles the UpdateTrip RPC.
func (h *Handler) UpdateTrip(ctx context.Context, req *connect.Request[tripv1.UpdateTripRequest]) (*connect.Response[tripv1.UpdateTripResponse], error) {
	TripID , err := uuid.Parse(req.Msg.TripId)
	if err != nil {
		return nil , connect.NewError(connect.CodeInvalidArgument,err)
	}

	trip, err := h.service.UpdateTrip(ctx, TripID, req.Msg.Status.String())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	driverIdStr := ""
	if trip.DriverID != nil {
		driverIdStr = trip.DriverID.String()
	}

	//json response 
	response := &tripv1.UpdateTripResponse{
		Trip: &tripv1.Trip{
			Id:       trip.ID.String(),
			UserId:   trip.PassengerID.String(),
			DriverId: driverIdStr,
			PickupLocation: &tripv1.Location{
				Latitude:  trip.PickupLat,
				Longitude: trip.PickupLng,
			},
			DropoffLocation: &tripv1.Location{
				Latitude:  trip.DropoffLat,
				Longitude: trip.DropoffLng,
			},
			Status: tripv1.TripStatus(tripv1.TripStatus_value[trip.Status]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: trip.CreatedAt.Unix(),
			},
			UpdatedAt: &timestamppb.Timestamp{
				Seconds: trip.UpdatedAt.Unix(),
			},
		},
	}
	
	return connect.NewResponse(response), nil
}

// CancelTrip handles the CancelTrip RPC.
func (h *Handler) CancelTrip(ctx context.Context, req *connect.Request[tripv1.CancelTripRequest]) (*connect.Response[tripv1.CancelTripResponse], error) {
	TripID , err := uuid.Parse(req.Msg.TripId)
	if err != nil {
		return nil , connect.NewError(connect.CodeInvalidArgument,err)
	}

	trip, err := h.service.CancelTrip(ctx, TripID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	driverIdStr := ""
	if trip.DriverID != nil {
		driverIdStr = trip.DriverID.String()
	}

	response := &tripv1.CancelTripResponse{
		Trip: &tripv1.Trip{
			Id:       trip.ID.String(),
			UserId:   trip.PassengerID.String(),
			DriverId: driverIdStr,
			PickupLocation: &tripv1.Location{
				Latitude:  trip.PickupLat,
				Longitude: trip.PickupLng,
			},
			DropoffLocation: &tripv1.Location{
				Latitude:  trip.DropoffLat,
				Longitude: trip.DropoffLng,
			},
			Status: tripv1.TripStatus(tripv1.TripStatus_value[trip.Status]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: trip.CreatedAt.Unix(),
			},
			UpdatedAt: &timestamppb.Timestamp{
				Seconds: trip.UpdatedAt.Unix(),
			},
		},
	}
	
	return connect.NewResponse(response), nil
}

// CompleteTrip handles the CompleteTrip RPC.
func (h *Handler) CompleteTrip(ctx context.Context, req *connect.Request[tripv1.CompleteTripRequest]) (*connect.Response[tripv1.CompleteTripResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("CompleteTrip not implemented"))
}
