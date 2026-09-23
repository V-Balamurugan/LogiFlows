// Package docs contains generated Swagger / OpenAPI documentation for LogiFlows API.
package docs

import (
	_ "embed"

	"github.com/swaggo/swag"
)

//go:embed swagger.json
var docTemplate string

// SwaggerInfo holds exported Swagger Info so clients can inspect or modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "LogiFlows API",
	Description:      "LogiFlows Intelligent End-to-End Logistics Coordination and Delivery Management System API.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
