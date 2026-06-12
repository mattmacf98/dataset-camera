package datasetcamera

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"os"

	"go.viam.com/rdk/app"
	camera "go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/camera/rtppassthrough"
	viamdata "go.viam.com/rdk/data"
	"go.viam.com/rdk/gostream"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

var DatasetCamera = resource.NewModel("mattmacf", "dataset-camera", "dataset-camera")

const binaryDataPageLimit = 100
const maxLimit = 1000

func init() {
	resource.RegisterComponent(camera.API, DatasetCamera,
		resource.Registration[camera.Camera, *Config]{
			Constructor: newDatasetCameraDatasetCamera,
		},
	)
}

type Config struct {
	DatasetID string `json:"dataset_id"`
	APIKey    string `json:"api_key,omitempty"`
	APIKeyID  string `json:"api_key_id,omitempty"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.DatasetID == "" {
		return nil, nil, fmt.Errorf("%s: dataset_id is required", path)
	}
	return nil, nil, nil
}

type binaryEntry struct {
	id        string
	timestamp time.Time
}

type datasetCameraDatasetCamera struct {
	resource.AlwaysRebuild
	resource.Named

	name resource.Name

	logger logging.Logger
	cfg    *Config

	viamClient *app.ViamClient
	dataClient *app.DataClient

	binaryEntries []binaryEntry

	apiKey   string
	apiKeyID string

	imageMu    sync.Mutex
	imageIndex int

	cancelCtx  context.Context
	cancelFunc func()
}

func newDatasetCameraDatasetCamera(
	ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger,
) (camera.Camera, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDatasetCamera(ctx, deps, rawConf.ResourceName(), conf, logger)
}

func NewDatasetCamera(
	ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger,
) (camera.Camera, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	apiKey := conf.APIKey
	apiKeyID := conf.APIKeyID
	if apiKey == "" || apiKeyID == "" {
		logger.Warnf("VIAM_API_KEY and VIAM_API_KEY_ID must be set in config, fallback to env")
		apiKey = os.Getenv("VIAM_API_KEY")
		apiKeyID = os.Getenv("VIAM_API_KEY_ID")
	}

	viamClient, err := app.CreateViamClientWithAPIKey(ctx, app.Options{}, apiKey, apiKeyID, logger)
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to create viam client: %w", err)
	}

	dataClient := viamClient.DataClient()
	entries, err := loadDatasetBinaryIDs(ctx, dataClient, conf.DatasetID, logger)
	if err != nil {
		_ = viamClient.Close()
		cancelFunc()
		return nil, fmt.Errorf("failed to load dataset binary IDs: %w", err)
	}

	logger.Infof("loaded %d binary data entries from dataset %s", len(entries), conf.DatasetID)

	return &datasetCameraDatasetCamera{
		name:          name,
		logger:        logger,
		cfg:           conf,
		viamClient:    viamClient,
		dataClient:    dataClient,
		apiKey:        apiKey,
		apiKeyID:      apiKeyID,
		binaryEntries: entries,
		cancelCtx:     cancelCtx,
		cancelFunc:    cancelFunc,
	}, nil
}

func loadDatasetBinaryIDs(ctx context.Context, dataClient *app.DataClient, datasetID string, logger logging.Logger) ([]binaryEntry, error) {
	filter := &app.Filter{DatasetID: datasetID}
	var entries []binaryEntry
	last := ""

	for {
		if len(entries) >= maxLimit {
			break
		}

		logger.Infof("getting binary data by filter: %v", filter)
		resp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
			Filter:              filter,
			Limit:               binaryDataPageLimit,
			Last:                last,
			IncludeInternalData: true,
		})
		if err != nil {
			logger.Errorf("failed to get binary data by filter: %v", err)
			return nil, err
		}
		logger.Infof("got %v binary data entries from dataset %s", resp, datasetID)
		if len(resp.BinaryData) == 0 {
			break
		}

		for _, data := range resp.BinaryData {
			if data == nil || data.Metadata == nil {
				continue
			}
			id := data.Metadata.BinaryDataID
			if id == "" {
				id = data.Metadata.ID
			}
			if id == "" {
				continue
			}
			entries = append(entries, binaryEntry{
				id:        id,
				timestamp: data.Metadata.TimeReceived,
			})
		}

		if resp.Last == "" || resp.Last == last {
			break
		}
		last = resp.Last
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	return entries, nil
}

func (s *datasetCameraDatasetCamera) Name() resource.Name {
	return s.name
}

func (s *datasetCameraDatasetCamera) Stream(ctx context.Context, errHandlers ...gostream.ErrorHandler) (gostream.VideoStream, error) {
	var videoStreamRetVal gostream.VideoStream

	return videoStreamRetVal, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Images(
	ctx context.Context, filterSourceNames []string, extra map[string]interface{},
) ([]camera.NamedImage, resource.ResponseMetadata, error) {
	if len(s.binaryEntries) == 0 {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("dataset %s has no binary data", s.cfg.DatasetID)
	}

	s.imageMu.Lock()
	entry := s.binaryEntries[s.imageIndex%len(s.binaryEntries)]
	s.imageIndex++
	s.imageMu.Unlock()

	binaryData, err := s.dataClient.BinaryDataByIDs(ctx, []string{entry.id}, &app.BinaryDataByIDsOptions{
		IncludeBinary: true,
	})
	if err != nil {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("failed to download binary data %s: %w", entry.id, err)
	}
	if len(binaryData) == 0 || binaryData[0] == nil {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("no binary data returned for id %s", entry.id)
	}

	binData := binaryData[0]
	if len(binData.Binary) == 0 {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("binary data %s has no image bytes", entry.id)
	}

	mimeType := ""
	if binData.Metadata != nil {
		mimeType = binData.Metadata.CaptureMetadata.MimeType
	}

	sourceName := s.name.Name
	namedImage, err := camera.NamedImageFromBytes(binData.Binary, sourceName, mimeType, viamdata.Annotations{})
	if err != nil {
		return nil, resource.ResponseMetadata{}, fmt.Errorf("failed to construct named image: %w", err)
	}

	capturedAt := entry.timestamp
	if binData.Metadata != nil && !binData.Metadata.TimeReceived.IsZero() {
		capturedAt = binData.Metadata.TimeReceived
	}

	return []camera.NamedImage{namedImage}, resource.ResponseMetadata{CapturedAt: capturedAt, Attributes: map[string]interface{}{
		"dataset_id": s.cfg.DatasetID,
		"latitude":   12,
		"longitude":  50,
		"altitude":   100,
	}}, nil
}

func (s *datasetCameraDatasetCamera) NextPointCloud(ctx context.Context, extra map[string]interface{}) (pointcloud.PointCloud, error) {
	var pointCloudRetVal pointcloud.PointCloud

	return pointCloudRetVal, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Properties(ctx context.Context) (camera.Properties, error) {
	var propertiesRetVal camera.Properties

	return propertiesRetVal, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Status(ctx context.Context) (map[string]interface{}, error) {
	s.imageMu.Lock()
	currentIndex := s.imageIndex
	s.imageMu.Unlock()

	return map[string]interface{}{
		"dataset_id":  s.cfg.DatasetID,
		"image_count": len(s.binaryEntries),
		"image_index": currentIndex,
	}, nil
}

func (s *datasetCameraDatasetCamera) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Geometries(ctx context.Context, extra map[string]interface{}) ([]spatialmath.Geometry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) SubscribeRTP(
	ctx context.Context, bufferSize int, packetsCB rtppassthrough.PacketCallback,
) (rtppassthrough.Subscription, error) {
	var subscriptionRetVal rtppassthrough.Subscription

	return subscriptionRetVal, fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Unsubscribe(ctx context.Context, id rtppassthrough.SubscriptionID) error {
	return fmt.Errorf("not implemented")
}

func (s *datasetCameraDatasetCamera) Close(context.Context) error {
	s.cancelFunc()
	if s.viamClient != nil {
		return s.viamClient.Close()
	}
	return nil
}
