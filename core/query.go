package core

import (
	"fmt"
	"github.com/proskenion/proskenion/core/model"
)

var (
	ErrQueryProcessorQueryEmptyBlockchain          = fmt.Errorf("Failed QueryProcessor Query blockchain is empty")
	ErrQueryProcessorQueryObjectCodeNotImplemented = fmt.Errorf("Failed QueryProcessor Query ObjectCode is not implemented")
	ErrQueryProcessorNotFound                      = fmt.Errorf("Failed QueryProcessor Query Not Found")
	ErrQueryProcessorNotExistAuthoirizer           = fmt.Errorf("Failed QueryProcessor No exists authorizer")
	ErrQueryProcessorNotSignedAuthorizer           = fmt.Errorf("Failed QueryProcessor Query don't sign authorizer")
)

type QueryProcessor interface {
	Query(wsv model.ObjectFinder, query model.Query) (model.QueryResponse, error)
}

type QueryValidator interface {
	Validate(wsv model.ObjectFinder, query model.Query) error
}

type QueryVerifier interface {
	Verify(query model.Query) error
}

type Querycutor interface {
	QueryProcessor
	QueryValidator
	QueryVerifier
}
