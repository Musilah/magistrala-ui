// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import "github.com/absmach/magistrala/pkg/errors"

var (
	errAuthorization          = errors.New("missing or invalid credentials provided")
	errAuthentication         = errors.New("failed to perform authentication over the entity")
	errMissingSecret          = errors.New("missing secret")
	errMissingIdentity        = errors.New("missing entity identity")
	errLimitSize              = errors.New("invalid limit size")
	errMissingConfigID        = errors.New("missing config id")
	errPageSize               = errors.New("invalid page size")
	errMissingEmail           = errors.New("missing email")
	errMissingName            = errors.New("missing name")
	errMissingPassword        = errors.New("missing password")
	errMissingMetadata        = errors.New("missing entity metadata")
	errMissingConfirmPassword = errors.New("missing confirm password")
	errInvalidResetPassword   = errors.New("invalid reset password")
	errNameSize               = errors.New("invalid name size")
	errBearerKey              = errors.New("missing or invalid bearer entity key")
	errMissingThingID         = errors.New("missing thing id")
	errMissingItem            = errors.New("missing item")
	errMissingChannelID       = errors.New("missing channel id")
	errMissingUserID          = errors.New("missing user id")
	errMissingRelation        = errors.New("missing relation")
	errMissingGroupID         = errors.New("missing group id")
	errMissingParentID        = errors.New("missing parent id")
	errMissingDescription     = errors.New("missing description")
	errMissingThingKey        = errors.New("missing thing key")
	errMissingExternalID      = errors.New("missing external id")
	errMissingExternalKey     = errors.New("missing external key")
	errMissingChannel         = errors.New("missing channel")
	errMissingPayload         = errors.New("missing payload")
	errMissingError           = errors.New("missing error")
	errMissingRefreshToken    = errors.New("missing refresh token")
	errMissingRef             = errors.New("missing ref")
	errInvalidQueryParams     = errors.New("invalid query parameters")
	errFileFormat             = errors.New("invalid file format")
	errInvalidFile            = errors.New("unsupported file type")
	errMissingDomainID        = errors.New("missing domain id")
	errMissingRole            = errors.New("missing role")
	errMissingValue           = errors.New("missing value")
	errCookieDecryption       = errors.New("failed to decrypt the cookie")
)
