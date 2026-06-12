package main

import (
	"datasetcamera"
	dumbyvs "datasetcamera/dumby-vs"

	camera "go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	vision "go.viam.com/rdk/services/vision"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{camera.API, datasetcamera.DatasetCamera}, resource.APIModel{vision.API, dumbyvs.DumbVs})

}
