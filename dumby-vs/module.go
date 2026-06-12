package dumbyvs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	vision "go.viam.com/rdk/services/vision"
	vis "go.viam.com/rdk/vision"
	"go.viam.com/rdk/vision/classification"
	objdet "go.viam.com/rdk/vision/objectdetection"
	"go.viam.com/rdk/vision/viscapture"
)

var (
	DumbVs           = resource.NewModel("mattmacf", "dataset-camera", "dumb-vs")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterService(vision.API, DumbVs,
		resource.Registration[vision.Service, *Config]{
			Constructor: newDatasetCameraDumbVs,
		},
	)
}

type Config struct {
	CameraName string `json:"camera_name"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns three values:
//  1. Required dependencies: other resources that must exist for this resource to work.
//  2. Optional dependencies: other resources that may exist but are not required.
//  3. An error if any Config fields are missing or invalid.
//
// The `path` parameter indicates
// where this resource appears in the machine's JSON configuration
// (for example, "components.0"). You can use it in error messages
// to indicate which resource has a problem.
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.CameraName == "" {
		return nil, nil, fmt.Errorf("%s: camera_name is required", path)
	}
	return []string{cfg.CameraName}, nil, nil
}

type datasetCameraDumbVs struct {
	resource.AlwaysRebuild
	resource.Named

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
	camera     camera.Camera
}

func newDatasetCameraDumbVs(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (vision.Service, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewDatasetCameraDumbVs(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewDatasetCameraDumbVs(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (vision.Service, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	camera, err := camera.FromProvider(deps, conf.CameraName)
	if err != nil {
		return nil, err
	}

	return &datasetCameraDumbVs{
		name:       name,
		logger:     logger,
		cfg:        conf,
		camera:     camera,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}, nil
}

func (s *datasetCameraDumbVs) Name() resource.Name {
	return s.name
}

// DetectionsFromCamera returns a list of detections from the next image from a specified camera using a configured detector.
func (s *datasetCameraDumbVs) DetectionsFromCamera(ctx context.Context, cameraName string, extra map[string]interface{}) ([]objdet.Detection, error) {
	return nil, fmt.Errorf("not implemented")
}

// Detections returns a list of detections from a given image using a configured detector.
func (s *datasetCameraDumbVs) Detections(ctx context.Context, img *camera.NamedImage, extra map[string]interface{}) ([]objdet.Detection, error) {
	return nil, fmt.Errorf("not implemented")
}

// ClassificationsFromCamera returns a list of classifications from the next image from a specified camera using a configured classifier.
func (s *datasetCameraDumbVs) ClassificationsFromCamera(ctx context.Context, cameraName string, n int, extra map[string]interface{}) (classification.Classifications, error) {
	var classificationsRetVal classification.Classifications

	return classificationsRetVal, fmt.Errorf("not implemented")
}

// Classifications returns a list of classifications from a given image using a configured classifier.
func (s *datasetCameraDumbVs) Classifications(ctx context.Context, img *camera.NamedImage, n int, extra map[string]interface{}) (classification.Classifications, error) {
	var classificationsRetVal classification.Classifications

	return classificationsRetVal, fmt.Errorf("not implemented")
}

// GetObjectPointClouds returns a list of 3D point cloud objects and metadata from the latest 3D camera image using a specified segmenter.
func (s *datasetCameraDumbVs) GetObjectPointClouds(ctx context.Context, cameraName string, extra map[string]interface{}) ([]*vis.Object, error) {
	return nil, fmt.Errorf("not implemented")
}

// properties
func (s *datasetCameraDumbVs) GetProperties(ctx context.Context, extra map[string]interface{}) (*vision.Properties, error) {
	return nil, fmt.Errorf("not implemented")
}

// CaptureAllFromCamera returns the next image, detections, classifications, and objects all together, given a camera name. Used for
// visualization.
func (s *datasetCameraDumbVs) CaptureAllFromCamera(ctx context.Context, cameraName string, captureOptions viscapture.CaptureOptions, extra map[string]interface{}) (viscapture.VisCapture, error) {
	var visCaptureRetVal viscapture.VisCapture

	return visCaptureRetVal, fmt.Errorf("not implemented")
}

func (s *datasetCameraDumbVs) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	images, md, err := s.camera.Images(ctx, nil, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	s.logger.Infof("got %d images, metadata: %v", len(images), md.Attributes)

	for attr, value := range md.Attributes {
		s.logger.Infof("attribute: %s, value: %v", attr, value)
	}
	return map[string]interface{}{
		"image_count": len(images),
		"metadata":    fmt.Sprintf("%v", md),
	}, nil
}

func (s *datasetCameraDumbVs) Status(ctx context.Context) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *datasetCameraDumbVs) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
