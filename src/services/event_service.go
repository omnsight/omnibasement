package services

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/omnsight/omnibasement/gen/base/v1"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"github.com/omnsight/omniscent-library/src/clients"
	"github.com/omnsight/omniscent-library/src/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService struct {
	base.UnimplementedEventServiceServer

	DBClient   *clients.ArangoDBClient
	Collection driver.Collection
}

func NewEventService(client *clients.ArangoDBClient) (*EventService, error) {
	// Create events collection
	ctx := context.Background()
	collection, err := client.GetCreateCollection(ctx, "events", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create events collection: %v", err)
	}
	logrus.Infof("✅ Initialized collection %s", collection.Name())

	collection.EnsurePersistentIndex(ctx, []string{"happened_at"}, &driver.EnsurePersistentIndexOptions{
		InBackground: true,
	})

	service := &EventService{
		DBClient:   client,
		Collection: collection,
	}
	return service, nil
}

func (s *EventService) GetEvent(ctx context.Context, req *base.GetEventRequest) (*base.GetEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting event with ID: %s", req.GetKey())

	// Read document from collection
	var event model.Event
	meta, err := s.Collection.ReadDocument(ctx, req.GetKey(), &event)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("event not found")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &base.GetEventResponse{Event: &event}, nil
}

func (s *EventService) CreateEvent(ctx context.Context, req *base.CreateEventRequest) (*base.CreateEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating event")

	// Create document in collection
	var event model.Event
	ctxWithReturnNew := driver.WithReturnNew(ctx, &event)
	meta, err := s.Collection.CreateDocument(ctxWithReturnNew, req.GetEvent())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &base.CreateEventResponse{Event: &event}, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, req *base.UpdateEventRequest) (*base.UpdateEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating event with Key: %s", req.GetEvent().GetKey())

	// Update document in collection
	var event model.Event
	ctxWithReturnNew := driver.WithReturnNew(ctx, &event)
	meta, err := s.Collection.UpdateDocument(ctxWithReturnNew, req.GetEvent().GetKey(), req.GetEvent())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetEvent().GetKey(),
			}).Info("event not found for update")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetEvent().GetKey(),
		}).Error("failed to update event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &base.UpdateEventResponse{Event: &event}, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, req *base.DeleteEventRequest) (*base.DeleteEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting event with Key: %s", req.GetKey())

	// Remove document from collection
	_, err := s.Collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("event not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &base.DeleteEventResponse{}, nil
}
