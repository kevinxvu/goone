package swaggerutil

import (
	"github.com/vuduongtp/go-core/pkg/server"
)

// SwaggOKResp success empty response
type SwaggOKResp struct{} //	@name	SwaggOKResp

// SwaggErrResp error empty response
type SwaggErrResp struct{} //	@name	SwaggErrResp

// SwaggErrDetailsResp model error response
type SwaggErrDetailsResp struct {
	server.ErrorResponse
} //	@name	SwaggErrDetailsResp
