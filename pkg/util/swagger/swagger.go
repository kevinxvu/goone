package swaggerutil

import (
	"github.com/kevinxvu/goone/pkg/server/apperr"
	"github.com/kevinxvu/goone/pkg/util/request"
)

// SwaggOKResp success empty response
type SwaggOKResp struct{} //	@name	SwaggOKResp

// SwaggErrResp error empty response
type SwaggErrResp struct{} //	@name	SwaggErrResp

// SwaggErrDetailsResp model error response
type SwaggErrDetailsResp struct {
	apperr.ErrorResponse
} //	@name	SwaggErrDetailsResp

// ListRequest holds data of listing request from react-admin
type ListRequest struct {
	request.ListRequest
}
