package resolution

import (
	"errors"

	"github.com/fstanis/screenresolution"
)

func GetResolutionLogic() (*screenresolution.Resolution, error) {
	resolution := screenresolution.GetPrimary()

	if resolution == nil {
		return resolution, errors.New("GetResolutionLogic Err")
	}

	return resolution, nil
}