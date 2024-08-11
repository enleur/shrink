// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.1-0.20240802201120-fdf32da8560e DO NOT EDIT.
package api

// PostShortenJSONBody defines parameters for PostShorten.
type PostShortenJSONBody struct {
	Url *string `json:"url,omitempty"`
}

// PostShortenJSONRequestBody defines body for PostShorten for application/json ContentType.
type PostShortenJSONRequestBody PostShortenJSONBody
