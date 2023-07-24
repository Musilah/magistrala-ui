// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ultravioletrs/mainflux-ui/ui"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	contentType = "text/html"
	staticDir   = "ui/web/static"
	protocol    = "http"
)

var (
	errMalformedData     = errors.New("malformed request data")
	errMalformedSubtopic = errors.New("malformed subtopic")
	errNoCookie          = errors.New("failed to read token cookie")
	errUnauthorized      = errors.New("failed to login")
	ErrAuthentication    = errors.New("failed to perform authentication over the entity")
	referer              = ""
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc ui.Service, redirect string, tracer opentracing.Tracer, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	r := bone.New()
	r.Get("/", kithttp.NewServer(
		kitot.TraceServer(tracer, "index")(indexEndpoint(svc)),
		decodeIndexRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/login", kithttp.NewServer(
		kitot.TraceServer(tracer, "login")(loginEndpoint(svc)),
		decodeLoginRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/login", kithttp.NewServer(
		kitot.TraceServer(tracer, "token")(tokenEndpoint(svc)),
		decodeTokenRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/refresh_token", kithttp.NewServer(
		kitot.TraceServer(tracer, "refresh_token")(refreshTokenEndpoint(svc)),
		decodeRefreshTokenRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/logout", kithttp.NewServer(
		kitot.TraceServer(tracer, "logout")(logoutEndpoint(svc)),
		decodeLogoutRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/password", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user_password")(updatePasswordEndpoint(svc)),
		decodePasswordUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/password", kithttp.NewServer(
		kitot.TraceServer(tracer, "show_update_password")(showUpdatePasswordEndpoint(svc)),
		decodePasswordReset,
		encodeResponse,
		opts...,
	))

	r.Post("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_user")(createUserEndpoint(svc)),
		decodeUserCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/users/bulk", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_users")(createUsersEndpoint(svc)),
		decodeUsersCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_users")(listUsersEndpoint(svc)),
		decodeListUsersRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/users/enabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "enable_user")(enableUserEndpoint(svc)),
		decodeUserStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/users/disabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "disable_user")(disableUserEndpoint(svc)),
		decodeUserStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/users/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_user")(viewUserEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/users/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user")(updateUserEndpoint(svc)),
		decodeUserUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/users/:id/tags", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user_tags")(updateUserTagsEndpoint(svc)),
		decodeUserTagsUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/users/:id/identity", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user_identity")(updateUserIdentityEndpoint(svc)),
		decodeUserIdentityUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_thing")(createThingEndpoint(svc)),
		decodeThingCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/things/bulk", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_things")(createThingsEndpoint(svc)),
		decodeThingsCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things")(listThingsEndpoint(svc)),
		decodeListThingsRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/things/enabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "enable_thing")(enableThingEndpoint(svc)),
		decodeThingStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/things/disabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "disable_thing")(disableThingEndpoint(svc)),
		decodeThingStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_thing")(viewThingEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing")(updateThingEndpoint(svc)),
		decodeThingUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id/tags", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing_tags")(updateThingTagsEndpoint(svc)),
		decodeThingTagsUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id/secret", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing_secret")(updateThingSecretEndpoint(svc)),
		decodeThingSecretUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id/owner", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing_owner")(updateThingOwnerEndpoint(svc)),
		decodeThingOwnerUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_channels_by_thing")(listChannelsByThingEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id/connect", kithttp.NewServer(
		kitot.TraceServer(tracer, "connect_Channel")(connectChannelEndpoint(svc)),
		decodeConnectChannel,
		encodeResponse,
		opts...,
	))

	r.Post("/disconnectChannel", kithttp.NewServer(
		kitot.TraceServer(tracer, "disconnect_channel")(disconnectChannelEndpoint(svc)),
		decodeDisconnectChannel,
		encodeResponse,
		opts...,
	))

	r.Post("/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_channel")(createChannelEndpoint(svc)),
		decodeChannelCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/channels/bulk", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_channels")(createChannelsEndpoint(svc)),
		decodeChannelsCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/channels/enabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "enable_channel")(enableChannelEndpoint(svc)),
		decodeChannelStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/channels/disabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "disable_channel")(disableChannelEndpoint(svc)),
		decodeChannelStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/channels/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_channel")(viewChannelEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/channels/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_channel")(updateChannelEndpoint(svc)),
		decodeChannelUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_channels")(listChannelsEndpoint(svc)),
		decodeListChannelsRequest,
		encodeResponse,
		opts...,
	))
	r.Post("/channels/:id/connectThing", kithttp.NewServer(
		kitot.TraceServer(tracer, "connect_thing")(connectThingEndpoint(svc)),
		decodeConnectThing,
		encodeResponse,
		opts...,
	))
	r.Post("/channels/:id/shareThing", kithttp.NewServer(
		kitot.TraceServer(tracer, "share_thing")(shareThingEndpoint(svc)),
		decodeShareThingRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/disconnectThing", kithttp.NewServer(
		kitot.TraceServer(tracer, "disconnect_thing")(disconnectThingEndpoint(svc)),
		decodeDisconnectThing,
		encodeResponse,
		opts...,
	))

	r.Post("/connect", kithttp.NewServer(
		kitot.TraceServer(tracer, "connect_things")(connectEndpoint(svc)),
		decodeConnect,
		encodeResponse,
		opts...,
	))

	r.Post("/disconnect", kithttp.NewServer(
		kitot.TraceServer(tracer, "disconnect_things")(disconnectEndpoint(svc)),
		decodeDisconnect,
		encodeResponse,
		opts...,
	))

	r.Get("/channels/:id/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things_by_channel")(listThingsByChannelEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/things_policies", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_things_policies")(listThingsPoliciesEndpoint(svc)),
		decodeListPoliciesRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/things_policies", kithttp.NewServer(
		kitot.TraceServer(tracer, "add_things_policy")(addThingsPolicyEndpoint(svc)),
		decodeAddThingsPolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/things_policies/update", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_things_policy")(updateThingsPolicyEndpoint(svc)),
		decodeUpdatePolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/things_policies/delete", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete_things_policy")(deleteThingsPolicyEndpoint(svc)),
		decodeDeleteThingsPolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_group")(createGroupEndpoint(svc)),
		decodeGroupCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_groups")(listGroupsEndpoint(svc)),
		decodeListGroupsRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/bulk", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_groups")(createGroupsEndpoint(svc)),
		decodeGroupsCreation,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/enabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "enable_group")(enableGroupEndpoint(svc)),
		decodeGroupStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/disabled", kithttp.NewServer(
		kitot.TraceServer(tracer, "disable_group")(disableGroupEndpoint(svc)),
		decodeGroupStatusUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_group")(viewGroupEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_group_members")(listGroupMembersEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_group")(updateGroupEndpoint(svc)),
		decodeGroupUpdate,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id/members", kithttp.NewServer(
		kitot.TraceServer(tracer, "assign")(assignEndpoint(svc)),
		decodeAssignRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id/unassign", kithttp.NewServer(
		kitot.TraceServer(tracer, "unassign")(unassignEndpoint(svc)),
		decodeUnassignRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/policies", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_policies")(listPoliciesEndpoint(svc)),
		decodeListPoliciesRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/policies", kithttp.NewServer(
		kitot.TraceServer(tracer, "add_policy")(addPolicyEndpoint(svc)),
		decodeAddPolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/policies/update", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_policy")(updatePolicyEndpoint(svc)),
		decodeUpdatePolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/policies/delete", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete_policy")(deletePolicyEndpoint(svc)),
		decodeDeletePolicyRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/messages", kithttp.NewServer(
		kitot.TraceServer(tracer, "publish")(publishMessageEndpoint(svc)),
		decodePublishRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/readmessages", kithttp.NewServer(
		kitot.TraceServer(tracer, "read_messages")(readMessageEndpoint(svc)),
		decodeReadMessageRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/readmessages", kithttp.NewServer(
		kitot.TraceServer(tracer, "ws_connection")(wsConnectionEndpoint(svc)),
		decodeWsConnectionRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/deleted", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_deleted_clients")(listDeletedClientsEndpoint(svc)),
		decodeListDeletedClientsRequest,
		encodeResponse,
		opts...,
	))

	r.GetFunc("/version", mainflux.Health("ui", instanceID))
	r.Handle("/metrics", promhttp.Handler())

	// Static file handler
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/*", fs)

	return r
}

func decodeIndexRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := indexReq{
		token: token,
	}

	return req, nil
}

func decodeLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	req := loginReq{}

	return req, nil
}

func decodePasswordReset(_ context.Context, r *http.Request) (interface{}, error) {
	req := PasswordResetReq{}

	return req, nil
}

func decodeTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	identity := r.PostFormValue("username")
	secret := r.PostFormValue("password")

	req := tokenReq{
		Identity: identity,
		Secret:   secret,
	}

	return req, nil
}

func decodeRefreshTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	c, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, errors.Wrap(errNoCookie, err)
		}
		return nil, err
	}
	req := refreshTokenReq{
		RefreshToken: c.Value,
		ref:          referer,
	}

	return req, nil
}

func decodeLogoutRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := sendMessageReq{}

	return req, nil
}

func decodeUserCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}
	var tags []string
	if err := json.Unmarshal([]byte(r.PostFormValue("tags")), &tags); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	credentials := sdk.Credentials{
		Identity: r.PostFormValue("identity"),
		Secret:   r.PostFormValue("secret"),
	}
	user := sdk.User{
		Name:        r.PostFormValue("name"),
		Credentials: credentials,
		Tags:        tags,
		Metadata:    meta,
	}

	req := createUserReq{
		token: token,
		user:  user,
	}

	return req, nil
}

func decodeUsersCreation(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("usersFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	names := []string{}
	emails := []string{}
	passwords := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				req := createUsersReq{
					token:     token,
					Names:     names,
					Emails:    emails,
					Passwords: passwords,
				}

				return req, nil
			}

			return nil, err
		}
		names = append(names, string(row[0]))
		emails = append(emails, string(row[1]))
		passwords = append(passwords, string(row[2]))
	}

}

func decodeListUsersRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listUsersReq{
		token: token,
	}

	return req, nil
}

func decodeView(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := viewResourceReq{
		token: token,
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeUserUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateUserReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateUserReq{
		token:    token,
		id:       bone.GetValue(r, "id"),
		Name:     data.Name,
		Metadata: data.Metadata,
	}

	return req, nil
}

func decodeUserTagsUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data struct {
		Tags []string `json:"tags"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateUserTagsReq{
		token: token,
		id:    bone.GetValue(r, "id"),
		Tags:  data.Tags,
	}

	return req, nil
}

func decodeUserIdentityUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	id := bone.GetValue(r, "id")

	var data updateUserIdentityReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateUserIdentityReq{
		token:    token,
		id:       id,
		Identity: data.Identity,
	}

	return req, nil
}

func decodePasswordUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := updateUserPasswordReq{
		token:   token,
		OldPass: r.PostFormValue("oldpass"),
		NewPass: r.PostFormValue("newpass"),
	}

	return req, nil
}

func decodeUserStatusUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := updateUserStatusReq{
		token:  token,
		UserID: r.PostFormValue("userID"),
	}

	return req, nil
}

func getAuthorization(r *http.Request) (string, error) {
	referer = r.URL.String()
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return "", errors.Wrap(errNoCookie, err)
		}
		return "", err
	}

	return c.Value, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)
	ar, _ := response.(uiRes)

	for k, v := range ar.Headers() {
		w.Header().Set(k, v)
	}

	// Add cookies to the response header
	for _, cookie := range ar.Cookies() {
		http.SetCookie(w, cookie)
	}

	w.WriteHeader(ar.Code())

	if ar.Empty() {
		return nil
	}
	_, err := w.Write(ar.html)
	if err != nil {
		return err
	}

	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch true {
	case errors.Contains(err, errNoCookie),
		errors.Contains(err, errUnauthorized):
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
	case errors.Contains(err, errMalformedData),
		errors.Contains(err, errMalformedSubtopic):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, ui.ErrUnauthorizedAccess):
		w.WriteHeader(http.StatusForbidden)
	case errors.Contains(err, ErrAuthentication):
		w.Header().Set("Location", "/refresh_token")
		w.WriteHeader(http.StatusSeeOther)
	default:
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func decodeThingCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}
	var tags []string
	if err := json.Unmarshal([]byte(r.PostFormValue("tags")), &tags); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	credentials := sdk.Credentials{
		Identity: r.PostFormValue("identity"),
		Secret:   r.PostFormValue("secret"),
	}
	thing := sdk.Thing{
		Name:        r.PostFormValue("name"),
		ID:          r.PostFormValue("thingID"),
		Credentials: credentials,
		Tags:        tags,
		Metadata:    meta,
	}
	req := createThingReq{
		token: token,
		thing: thing,
	}

	return req, nil
}

func decodeListThingsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listThingsReq{
		token: token,
	}

	return req, nil
}

func decodeThingUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateThingReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateThingReq{
		token:    token,
		id:       bone.GetValue(r, "id"),
		Name:     data.Name,
		Metadata: data.Metadata,
	}

	return req, nil
}

func decodeThingTagsUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateThingTagsReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateThingTagsReq{
		token: token,
		id:    bone.GetValue(r, "id"),
		Tags:  data.Tags,
	}

	return req, nil
}

func decodeThingSecretUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateThingSecretReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateThingSecretReq{
		token:  token,
		id:     bone.GetValue(r, "id"),
		Secret: data.Secret,
	}

	return req, nil
}

func decodeThingStatusUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := updateThingStatusReq{
		token:   token,
		ThingID: r.PostFormValue("thingID"),
	}

	return req, nil
}

func decodeThingOwnerUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateThingOwnerReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateThingOwnerReq{
		token: token,
		id:    bone.GetValue(r, "id"),
		Owner: data.Owner,
	}

	return req, nil
}

func decodeThingsCreation(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("thingsFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	names := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				req := createThingsReq{
					token: token,
					Names: names,
				}
				return req, nil
			}
			return nil, err
		}
		names = append(names, string(row[0]))
	}
}

func decodeChannelCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	ch := sdk.Channel{
		Name:        r.PostFormValue("name"),
		Description: r.PostFormValue("description"),
		Metadata:    meta,
		ParentID:    r.PostFormValue("parentID"),
	}

	req := createChannelReq{
		token:   token,
		Channel: ch,
	}

	return req, nil
}

func decodeChannelsCreation(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("channelsFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	names := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				req := createChannelsReq{
					token: token,
					Names: names,
				}
				return req, nil
			}
			return nil, err
		}
		names = append(names, string(row[0]))
	}
}

func decodeChannelUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateChannelReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateChannelReq{
		token:       token,
		id:          bone.GetValue(r, "id"),
		Name:        data.Name,
		Metadata:    data.Metadata,
		Description: data.Description,
	}

	return req, nil
}

func decodeListChannelsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listChannelsReq{
		token: token,
	}

	return req, nil
}

func decodeConnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	chanID := bone.GetValue(r, "id")
	thingID := r.Form.Get("thingID")
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	connIDs := sdk.ConnectionIDs{
		ChannelIDs: []string{chanID},
		ThingIDs:   []string{thingID},
		Actions:    r.PostForm["actions"],
	}
	req := connectThingReq{
		token:   token,
		ConnIDs: connIDs,
	}

	return req, nil
}

func decodeShareThingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	chanID := bone.GetValue(r, "id")
	userID := r.Form.Get("userID")
	actions := r.PostForm["actions"]
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := shareThingReq{
		token:   token,
		ChanID:  chanID,
		UserID:  userID,
		Actions: actions,
	}
	return req, nil

}

func decodeConnectChannel(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	chanID := r.Form.Get("channelID")
	thingID := bone.GetValue(r, "id")
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	connIDs := sdk.ConnectionIDs{
		ChannelIDs: []string{chanID},
		ThingIDs:   []string{thingID},
		Actions:    r.PostForm["actions"],
	}
	req := connectChannelReq{
		token:   token,
		ConnIDs: connIDs,
	}

	return req, nil
}

func decodeConnect(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("thingsFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	chanID := r.Form.Get("chanID")

	chanIDs := []string{}
	thingIDs := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				connIDs := sdk.ConnectionIDs{
					ChannelIDs: chanIDs,
					ThingIDs:   thingIDs,
				}
				req := connectReq{
					token:   token,
					ConnIDs: connIDs,
				}

				return req, nil
			}

			return nil, err
		}
		thingIDs = append(thingIDs, string(row[0]))
		chanIDs = append(chanIDs, chanID)
	}
}

func decodeDisconnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	chanID := r.Form.Get("channelID")
	thingID := r.Form.Get("thingID")
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := disconnectThingReq{
		token:   token,
		ChanID:  chanID,
		ThingID: thingID,
	}

	return req, nil
}

func decodeDisconnectChannel(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	chanID := r.Form.Get("channelID")
	thingID := r.Form.Get("thingID")
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := disconnectChannelReq{
		token:   token,
		ChanID:  chanID,
		ThingID: thingID,
	}

	return req, nil
}

func decodeDisconnect(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("thingsFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	chanID := r.Form.Get("chanID")

	chanIDs := []string{}
	thingIDs := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				connIDs := sdk.ConnectionIDs{
					ChannelIDs: chanIDs,
					ThingIDs:   thingIDs,
				}
				req := disconnectReq{
					token:   token,
					ConnIDs: connIDs,
				}

				return req, nil
			}

			return nil, err
		}
		thingIDs = append(thingIDs, string(row[0]))
		chanIDs = append(chanIDs, chanID)
	}
}

func decodeChannelStatusUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := updateChannelStatusReq{
		token:     token,
		ChannelID: r.PostFormValue("channelID"),
	}

	return req, nil
}

func decodeAddThingsPolicyRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	thingID := r.PostFormValue("subject")
	chanID := r.PostFormValue("object")

	connIDs := sdk.ConnectionIDs{
		ChannelIDs: []string{chanID},
		ThingIDs:   []string{thingID},
		Actions:    r.PostForm["actions"],
	}
	req := addThingsPolicyReq{
		token:   token,
		ConnIDs: connIDs,
	}
	return req, nil
}

func decodeDeleteThingsPolicyRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	thingID := r.PostFormValue("subject")
	chanID := r.PostFormValue("object")

	connIDs := sdk.ConnectionIDs{
		ChannelIDs: []string{chanID},
		ThingIDs:   []string{thingID},
		Actions:    r.PostForm["actions"],
	}
	req := deleteThingsPolicyReq{
		token:   token,
		ConnIDs: connIDs,
	}
	return req, nil
}

func decodeGroupCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	group := sdk.Group{
		Name:        r.PostFormValue("name"),
		Description: r.PostFormValue("description"),
		Metadata:    meta,
		ParentID:    r.PostFormValue("parentID"),
	}
	req := createGroupReq{
		token: token,
		Group: group,
	}

	return req, nil
}

func decodeGroupsCreation(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("groupsFile")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".csv") {
		return nil, errors.New("unsupported file type")
	}
	csvr := csv.NewReader(file)

	names := []string{}
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				req := createGroupsReq{
					token: token,
					Names: names,
				}
				return req, nil
			}
			return nil, err
		}
		names = append(names, string(row[0]))
	}
}

func decodeListGroupsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listGroupsReq{
		token: token,
	}

	return req, nil
}

func decodeGroupUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	var data updateGroupReq
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	req := updateGroupReq{
		token:       token,
		id:          bone.GetValue(r, "id"),
		Name:        data.Name,
		Metadata:    data.Metadata,
		Description: data.Description,
	}
	return req, nil
}

func decodeAssignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	memberid := r.PostFormValue("memberID")
	memberType := r.PostForm["Type"]

	req := assignReq{
		token:    token,
		groupID:  bone.GetValue(r, "id"),
		MemberID: memberid,
		Type:     memberType,
	}

	return req, nil
}

func decodeUnassignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	MemberId := r.PostFormValue("memberID")

	req := unassignReq{
		token:    token,
		groupID:  bone.GetValue(r, "id"),
		MemberID: MemberId,
	}

	return req, nil
}

func decodeGroupStatusUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := updateGroupStatusReq{
		token:   token,
		GroupID: r.PostFormValue("groupID"),
	}

	return req, nil
}

func decodeListPoliciesRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listPoliciesReq{
		token: token,
	}

	return req, nil
}

func decodeAddPolicyRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	policy := sdk.Policy{
		Subject: r.PostFormValue("subject"),
		Object:  r.PostFormValue("object"),
		Actions: r.PostForm["actions"],
	}

	req := addPolicyReq{
		token:  token,
		Policy: policy,
	}

	return req, nil
}

func decodeUpdatePolicyRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	var actions []string
	if err := json.Unmarshal([]byte(r.Form.Get("actions")), &actions); err != nil {
		return nil, err
	}

	policy := sdk.Policy{
		Subject: r.Form.Get("subject"),
		Object:  r.Form.Get("object"),
		Actions: actions,
	}

	req := updatePolicyReq{
		token:  token,
		Policy: policy,
	}

	return req, nil
}

func decodeDeletePolicyRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	policy := sdk.Policy{
		Object:  r.Form.Get("object"),
		Subject: r.Form.Get("subject"),
	}

	req := deletePolicyReq{
		token:  token,
		Policy: policy,
	}

	return req, nil
}

func decodePublishRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	msg := messaging.Message{
		Protocol: protocol,
		Channel:  r.Form.Get("channelID"),
		Subtopic: "",
		Payload:  []byte(r.Form.Get("message")),
		Created:  time.Now().UnixNano(),
	}

	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := publishReq{
		Msg:      &msg,
		thingKey: r.Form.Get("thingKey"),
		token:    token,
	}

	return req, nil
}

func decodeReadMessageRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	req := readMessageReq{
		token: token,
	}

	return req, nil
}

func decodeWsConnectionRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}

	chanID := r.Form.Get("chanID")
	thingKey := r.Form.Get("thingKey")

	req := wsConnectionReq{
		token:    token,
		ChanID:   chanID,
		ThingKey: thingKey,
	}

	return req, nil
}

func decodeListDeletedClientsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	token, err := getAuthorization(r)
	if err != nil {
		return nil, err
	}
	req := listDeletedClientsReq{
		token: token,
	}

	return req, nil
}