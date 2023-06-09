/*
 * Kontakt.io App API
 *
 * API to access and configure the Kontakt.io App
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ConfigurationApiController binds http requests to an api service and writes the service results to the http response
type ConfigurationApiController struct {
	service      ConfigurationApiServicer
	errorHandler ErrorHandler
}

// ConfigurationApiOption for how the controller is set up.
type ConfigurationApiOption func(*ConfigurationApiController)

// WithConfigurationApiErrorHandler inject ErrorHandler into controller
func WithConfigurationApiErrorHandler(h ErrorHandler) ConfigurationApiOption {
	return func(c *ConfigurationApiController) {
		c.errorHandler = h
	}
}

// NewConfigurationApiController creates a default api controller
func NewConfigurationApiController(s ConfigurationApiServicer, opts ...ConfigurationApiOption) Router {
	controller := &ConfigurationApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ConfigurationApiController
func (c *ConfigurationApiController) Routes() Routes {
	return Routes{
		{
			"GetConfigurations",
			strings.ToUpper("Get"),
			"/v1/configs",
			c.GetConfigurations,
		},
		{
			"PostConfiguration",
			strings.ToUpper("Post"),
			"/v1/configs",
			c.PostConfiguration,
		},
	}
}

// GetConfigurations - Get all Kontakt.io configurations
func (c *ConfigurationApiController) GetConfigurations(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.GetConfigurations(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PostConfiguration - Creates a configuration
func (c *ConfigurationApiController) PostConfiguration(w http.ResponseWriter, r *http.Request) {
	configurationParam := Configuration{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&configurationParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertConfigurationRequired(configurationParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.PostConfiguration(r.Context(), configurationParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
