package builder

import (
	"encoding/json"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/export"
	"github.com/pentops/j5/internal/structure"
)

func DescriptorFromSource(img *source_j5pb.SourceImage) (*schema_j5pb.API, error) {
	reflectAPI, err := structure.ReflectFromSource(img)
	if err != nil {
		return nil, err
	}

	descriptorAPI, err := reflectAPI.ToJ5Proto()
	if err != nil {
		return nil, err
	}

	err = structure.ResolveProse(img, descriptorAPI)
	if err != nil {
		return nil, err
	}

	return descriptorAPI, nil
}

func SwaggerFromDescriptor(descriptorAPI *schema_j5pb.API) ([]byte, error) {
	swaggerDoc, err := export.BuildSwagger(descriptorAPI)
	if err != nil {
		return nil, err
	}

	asJson, err := json.Marshal(swaggerDoc)
	if err != nil {
		return nil, err
	}

	return asJson, nil
}

func JDefFromDescriptor(descriptorAPI *schema_j5pb.API) ([]byte, error) {
	jDefJSON, err := export.FromProto(descriptorAPI)
	if err != nil {
		return nil, err
	}

	jDefJSONBytes, err := json.Marshal(jDefJSON)
	if err != nil {
		return nil, err
	}

	return jDefJSONBytes, nil
}
