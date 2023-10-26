// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	// CTJSON represents JSON content type.
	CTJSON ContentType = "application/json"

	// CTJSONSenML represents JSON SenML content type.
	CTJSONSenML ContentType = "application/senml+json"

	// CTBinary represents binary content type.
	CTBinary ContentType = "application/octet-stream"

	// EnabledStatus represents enable status for a client.
	EnabledStatus = "enabled"

	// DisabledStatus represents disabled status for a client.
	DisabledStatus = "disabled"

	BearerPrefix = "Bearer "

	ThingPrefix = "Thing "
)

// ContentType represents all possible content types.
type ContentType string

var _ SDK = (*mfSDK)(nil)

var (
	// ErrFailedCreation indicates that entity creation failed.
	ErrFailedCreation = errors.New("failed to create entity in the db")

	// ErrFailedList indicates that entities list failed.
	ErrFailedList = errors.New("failed to list entities")

	// ErrFailedUpdate indicates that entity update failed.
	ErrFailedUpdate = errors.New("failed to update entity")

	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.New("failed to fetch entity")

	// ErrFailedRemoval indicates that entity removal failed.
	ErrFailedRemoval = errors.New("failed to remove entity")

	// ErrFailedEnable indicates that client enable failed.
	ErrFailedEnable = errors.New("failed to enable client")

	// ErrFailedDisable indicates that client disable failed.
	ErrFailedDisable = errors.New("failed to disable client")

	ErrInvalidJWT = errors.New("invalid JWT")
)

type PageMetadata struct {
	Total      uint64   `json:"total"`
	Offset     uint64   `json:"offset"`
	Limit      uint64   `json:"limit"`
	Level      uint64   `json:"level,omitempty"`
	Email      string   `json:"email,omitempty"`
	Name       string   `json:"name,omitempty"`
	Type       string   `json:"type,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
	Status     string   `json:"status,omitempty"`
	Action     string   `json:"action,omitempty"`
	Subject    string   `json:"subject,omitempty"`
	Object     string   `json:"object,omitempty"`
	Permission string   `json:"permission,omitempty"`
	Tag        string   `json:"tag,omitempty"`
	Owner      string   `json:"owner,omitempty"`
	SharedBy   string   `json:"shared_by,omitempty"`
	Visibility string   `json:"visibility,omitempty"`
	OwnerID    string   `json:"owner_id,omitempty"`
	Topic      string   `json:"topic,omitempty"`
	Contact    string   `json:"contact,omitempty"`
	State      string   `json:"state,omitempty"`
}

// Credentials represent client credentials: it contains
// "identity" which can be a username, email, generated name;
// and "secret" which can be a password or access token.
type Credentials struct {
	Identity string `json:"identity,omitempty"` // username or generated login ID
	Secret   string `json:"secret,omitempty"`   // password or token
}

// SDK contains Mainflux API.
type SDK interface {
	// CreateUser registers mainflux user.
	//
	// example:
	//  user := sdk.User{
	//    Name:	 "John Doe",
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//      Secret:   "12345678",
	//    },
	//  }
	//  user, _ := sdk.CreateUser(user)
	//  fmt.Println(user)
	CreateUser(user User, token string) (User, errors.SDKError)

	// User returns user object by id.
	//
	// example:
	//  user, _ := sdk.User("userID", "token")
	//  fmt.Println(user)
	User(id, token string) (User, errors.SDKError)

	// Users returns list of users.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Name:   "John Doe",
	//	}
	//	users, _ := sdk.Users(pm, "token")
	//	fmt.Println(users)
	Users(pm PageMetadata, token string) (UsersPage, errors.SDKError)

	// UserProfile returns user logged in.
	//
	// example:
	//  user, _ := sdk.UserProfile("token")
	//  fmt.Println(user)
	UserProfile(token string) (User, errors.SDKError)

	// UpdateUser updates existing user.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "userID",
	//    Name: "John Doe",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  user, _ := sdk.UpdateUser(user, "token")
	//  fmt.Println(user)
	UpdateUser(user User, token string) (User, errors.SDKError)

	// UpdateUserTags updates the user's tags.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "userID",
	//    Tags: []string{"tag1", "tag2"},
	//  }
	//  user, _ := sdk.UpdateUserTags(user, "token")
	//  fmt.Println(user)
	UpdateUserTags(user User, token string) (User, errors.SDKError)

	// UpdateUserIdentity updates the user's identity
	//
	// example:
	//  user := sdk.User{
	//    ID:   "userID",
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//    },
	//  }
	//  user, _ := sdk.UpdateUserIdentity(user, "token")
	//  fmt.Println(user)
	UpdateUserIdentity(user User, token string) (User, errors.SDKError)

	// UpdateUserOwner updates the user's owner.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "userID",
	//    Owner: "ownerID",
	//  }
	//  user, _ := sdk.UpdateUserOwner(user, "token")
	//  fmt.Println(user)
	UpdateUserOwner(user User, token string) (User, errors.SDKError)

	// ResetPasswordRequest sends a password request email to a user.
	//
	// example:
	//  err := sdk.ResetPasswordRequest("example@email.com")
	//  fmt.Println(err)
	ResetPasswordRequest(email string) errors.SDKError

	// ResetPassword changes a user's password to the one passed in the argument.
	//
	// example:
	//  err := sdk.ResetPassword("password","password","token")
	//  fmt.Println(err)
	ResetPassword(password, confPass, token string) errors.SDKError

	// UpdatePassword updates user password.
	//
	// example:
	//  user, _ := sdk.UpdatePassword("oldPass", "newPass", "token")
	//  fmt.Println(user)
	UpdatePassword(oldPass, newPass, token string) (User, errors.SDKError)

	// EnableUser changes the status of the user to enabled.
	//
	// example:
	//  user, _ := sdk.EnableUser("userID", "token")
	//  fmt.Println(user)
	EnableUser(id, token string) (User, errors.SDKError)

	// DisableUser changes the status of the user to disabled.
	//
	// example:
	//  user, _ := sdk.DisableUser("userID", "token")
	//  fmt.Println(user)
	DisableUser(id, token string) (User, errors.SDKError)

	// CreateToken receives credentials and returns user token.
	//
	// example:
	//  user := sdk.User{
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//      Secret:   "12345678",
	//    },
	//  }
	//  token, _ := sdk.CreateToken(user)
	//  fmt.Println(token)
	CreateToken(user User) (Token, errors.SDKError)

	// RefreshToken receives credentials and returns user token.
	//
	// example:
	//  token, _ := sdk.RefreshToken("refresh_token")
	//  fmt.Println(token)
	RefreshToken(token string) (Token, errors.SDKError)

	// ListUserChannels list all channels belongs a particular user id.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  channels, _ := sdk.ListUserChannels("user_id_1", pm, "token")
	//  fmt.Println(channels)
	ListUserChannels(userID string, pm PageMetadata, token string) (ChannelsPage, errors.SDKError)

	// ListUserGroups list all groups belongs a particular user id.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  groups, _ := sdk.ListUserGroups("user_id_1", pm, "token")
	//  fmt.Println(channels)
	ListUserGroups(userID string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// ListUserThings list all things belongs a particular user id.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  things, _ := sdk.ListUserThings("user_id_1", pm, "token")
	//  fmt.Println(things)
	ListUserThings(userID string, pm PageMetadata, token string) (ThingsPage, errors.SDKError)

	// CreateThing registers new thing and returns its id.
	//
	// example:
	//  thing := sdk.Thing{
	//    Name: "My Thing",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  thing, _ := sdk.CreateThing(thing, "token")
	//  fmt.Println(thing)
	CreateThing(thing Thing, token string) (Thing, errors.SDKError)

	// CreateThings registers new things and returns their ids.
	//
	// example:
	//  things := []sdk.Thing{
	//    {
	//      Name: "My Thing 1",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//    {
	//      Name: "My Thing 2",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//  }
	//  things, _ := sdk.CreateThings(things, "token")
	//  fmt.Println(things)
	CreateThings(things []Thing, token string) ([]Thing, errors.SDKError)

	// Filters things and returns a page result.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Thing",
	//  }
	//  things, _ := sdk.Things(pm, "token")
	//  fmt.Println(things)
	Things(pm PageMetadata, token string) (ThingsPage, errors.SDKError)

	// ThingsByChannel returns page of things that are connected to specified channel.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Thing",
	//  }
	//  things, _ := sdk.ThingsByChannel("channelID", pm, "token")
	//  fmt.Println(things)
	ThingsByChannel(chanID string, pm PageMetadata, token string) (ThingsPage, errors.SDKError)

	// Thing returns thing object by id.
	//
	// example:
	//  thing, _ := sdk.Thing("thingID", "token")
	//  fmt.Println(thing)
	Thing(id, token string) (Thing, errors.SDKError)

	// UpdateThing updates existing thing.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:   "thingID",
	//    Name: "My Thing",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  thing, _ := sdk.UpdateThing(thing, "token")
	//  fmt.Println(thing)
	UpdateThing(thing Thing, token string) (Thing, errors.SDKError)

	// UpdateThingTags updates the client's tags.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:   "thingID",
	//    Tags: []string{"tag1", "tag2"},
	//  }
	//  thing, _ := sdk.UpdateThingTags(thing, "token")
	//  fmt.Println(thing)
	UpdateThingTags(thing Thing, token string) (Thing, errors.SDKError)

	// UpdateThingSecret updates the client's secret
	//
	// example:
	//  thing, err := sdk.UpdateThingSecret("thingID", "newSecret", "token")
	//  fmt.Println(thing)
	UpdateThingSecret(id, secret, token string) (Thing, errors.SDKError)

	// UpdateThingOwner updates the client's owner.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:    "thingID",
	//    Owner: "ownerID",
	//  }
	//  thing, _ := sdk.UpdateThingOwner(thing, "token")
	//  fmt.Println(thing)
	UpdateThingOwner(thing Thing, token string) (Thing, errors.SDKError)

	// EnableThing changes client status to enabled.
	//
	// example:
	//  thing, _ := sdk.EnableThing("thingID", "token")
	//  fmt.Println(thing)
	EnableThing(id, token string) (Thing, errors.SDKError)

	// DisableThing changes client status to disabled - soft delete.
	//
	// example:
	//  thing, _ := sdk.DisableThing("thingID", "token")
	//  fmt.Println(thing)
	DisableThing(id, token string) (Thing, errors.SDKError)

	// IdentifyThing validates thing's key and returns its ID
	//
	// example:
	//  id, _ := sdk.IdentifyThing("thingKey")
	//  fmt.Println(id)
	IdentifyThing(key string) (string, errors.SDKError)

	// ShareThing shares thing with other users.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	//  	UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	//  err := sdk.ShareThing("thing_id", req, "token")
	//  fmt.Println(err)
	ShareThing(thingID string, req UsersRelationRequest, token string) errors.SDKError

	// UnshareThing unshare a thing with other users.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	//  	UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	//  err := sdk.UnshareThing("thing_id", req, "token")
	//  fmt.Println(err)
	UnshareThing(thingID string, req UsersRelationRequest, token string) errors.SDKError

	// ListThingUsers all users in a thing.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  users, _ := sdk.ListThingUsers("thing_id", pm, "token")
	//  fmt.Println(users)
	ListThingUsers(thingID string, pm PageMetadata, token string) (UsersPage, errors.SDKError)

	// CreateGroup creates new group and returns its id.
	//
	// example:
	//  group := sdk.Group{
	//    Name: "My Group",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  group, _ := sdk.CreateGroup(group, "token")
	//  fmt.Println(group)
	CreateGroup(group Group, token string) (Group, errors.SDKError)

	// Groups returns page of groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Groups(pm, "token")
	//  fmt.Println(groups)
	Groups(pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Parents returns page of users groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Parents("groupID", pm, "token")
	//  fmt.Println(groups)
	Parents(id string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Children returns page of users groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Children("groupID", pm, "token")
	//  fmt.Println(groups)
	Children(id string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Group returns users group object by id.
	//
	// example:
	//  group, _ := sdk.Group("groupID", "token")
	//  fmt.Println(group)
	Group(id, token string) (Group, errors.SDKError)

	// UpdateGroup updates existing group.
	//
	// example:
	//  group := sdk.Group{
	//    ID:   "groupID",
	//    Name: "My Group",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  group, _ := sdk.UpdateGroup(group, "token")
	//  fmt.Println(group)
	UpdateGroup(group Group, token string) (Group, errors.SDKError)

	// EnableGroup changes group status to enabled.
	//
	// example:
	//  group, _ := sdk.EnableGroup("groupID", "token")
	//  fmt.Println(group)
	EnableGroup(id, token string) (Group, errors.SDKError)

	// DisableGroup changes group status to disabled - soft delete.
	//
	// example:
	//  group, _ := sdk.DisableGroup("groupID", "token")
	//  fmt.Println(group)
	DisableGroup(id, token string) (Group, errors.SDKError)

	// AddUserToGroup add user to a group.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	//  	UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	// group, _ := sdk.AddUserToGroup("groupID",req, "token")
	// fmt.Println(group)
	AddUserToGroup(groupID string, req UsersRelationRequest, token string) errors.SDKError

	// RemoveUserFromGroup remove user from a group.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	//  	UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	// group, _ := sdk.RemoveUserFromGroup("groupID",req, "token")
	// fmt.Println(group)
	RemoveUserFromGroup(groupID string, req UsersRelationRequest, token string) errors.SDKError

	// ListGroupUsers list all users in the group id .
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  groups, _ := sdk.ListGroupUsers("groupID", pm, "token")
	//  fmt.Println(groups)
	ListGroupUsers(groupID string, pm PageMetadata, token string) (UsersPage, errors.SDKError)

	// ListGroupChannels list all channels in the group id .
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit", // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  groups, _ := sdk.ListGroupChannels("groupID", pm, "token")
	//  fmt.Println(groups)
	ListGroupChannels(groupID string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// CreateChannel creates new channel and returns its id.
	//
	// example:
	//  channel := sdk.Channel{
	//    Name: "My Channel",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  channel, _ := sdk.CreateChannel(channel, "token")
	//  fmt.Println(channel)
	CreateChannel(channel Channel, token string) (Channel, errors.SDKError)

	// CreateChannels registers new channels and returns their ids.
	//
	// example:
	//  channels := []sdk.Channel{
	//    {
	//      Name: "My Channel 1",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//    {
	//      Name: "My Channel 2",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//  }
	//  channels, _ := sdk.CreateChannels(channels, "token")
	//  fmt.Println(channels)
	CreateChannels(channels []Channel, token string) ([]Channel, errors.SDKError)

	// Channels returns page of channels.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Channel",
	//  }
	//  channels, _ := sdk.Channels(pm, "token")
	//  fmt.Println(channels)
	Channels(pm PageMetadata, token string) (ChannelsPage, errors.SDKError)

	// ChannelsByThing returns page of channels that are connected to specified thing.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Channel",
	//  }
	//  channels, _ := sdk.ChannelsByThing("thingID", pm, "token")
	//  fmt.Println(channels)
	ChannelsByThing(thingID string, pm PageMetadata, token string) (ChannelsPage, errors.SDKError)

	// Channel returns channel data by id.
	//
	// example:
	//  channel, _ := sdk.Channel("channelID", "token")
	//  fmt.Println(channel)
	Channel(id, token string) (Channel, errors.SDKError)

	// UpdateChannel updates existing channel.
	//
	// example:
	//  channel := sdk.Channel{
	//    ID:   "channelID",
	//    Name: "My Channel",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  channel, _ := sdk.UpdateChannel(channel, "token")
	//  fmt.Println(channel)
	UpdateChannel(channel Channel, token string) (Channel, errors.SDKError)

	// EnableChannel changes channel status to enabled.
	//
	// example:
	//  channel, _ := sdk.EnableChannel("channelID", "token")
	//  fmt.Println(channel)
	EnableChannel(id, token string) (Channel, errors.SDKError)

	// DisableChannel changes channel status to disabled - soft delete.
	//
	// example:
	//  channel, _ := sdk.DisableChannel("channelID", "token")
	//  fmt.Println(channel)
	DisableChannel(id, token string) (Channel, errors.SDKError)

	// AddUserToChannel add user to a channel.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	// 		UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	// err := sdk.AddUserToChannel("channel_id", req, "token")
	// fmt.Println(err)
	AddUserToChannel(channelID string, req UsersRelationRequest, token string) errors.SDKError

	// RemoveUserFromChannel remove user from a group.
	//
	// example:
	// req := sdk.UsersRelationRequest{
	//		Relation: "viewer", // available options: "owner", "admin", "editor", "viewer"
	//  	UserIDs: ["user_id_1", "user_id_2", "user_id_3"]
	// }
	// err := sdk.RemoveUserFromChannel("channel_id", req, "token")
	// fmt.Println(err)
	RemoveUserFromChannel(channelID string, req UsersRelationRequest, token string) errors.SDKError

	// ListChannelUsers list all users in a channel .
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "edit",  // available Options:  "administrator", "delete", edit", "view", "share", "owner", "admin", "editor", "viewer"
	//	}
	//  users, _ := sdk.ListChannelUsers("channel_id", pm, "token")
	//  fmt.Println(users)
	ListChannelUsers(channelID string, pm PageMetadata, token string) (UsersPage, errors.SDKError)

	// AddUserGroupToChannel add user group to a channel.
	//
	// example:
	// req := sdk.UserGroupsRequest{
	//  	GroupsIDs: ["group_id_1", "group_id_2", "group_id_3"]
	// }
	// err := sdk.AddUserGroupToChannel("channel_id",req, "token")
	// fmt.Println(err)
	AddUserGroupToChannel(channelID string, req UserGroupsRequest, token string) errors.SDKError

	// RemoveUserGroupFromChannel remove user group from a channel.
	//
	// example:
	// req := sdk.UserGroupsRequest{
	//  	GroupsIDs: ["group_id_1", "group_id_2", "group_id_3"]
	// }
	// err := sdk.RemoveUserGroupFromChannel("channel_id",req, "token")
	// fmt.Println(err)
	RemoveUserGroupFromChannel(channelID string, req UserGroupsRequest, token string) errors.SDKError

	// ListChannelUserGroups list all user groups in a channel.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Permission: "view",
	//	}
	//  groups, _ := sdk.ListChannelUserGroups("channel_id_1", pm, "token")
	//  fmt.Println(groups)
	ListChannelUserGroups(channelID string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Connect bulk connects things to channels specified by id.
	//
	// example:
	//  conns := sdk.Connection{
	//    ChannelID: "channel_id_1",
	//    ThingID:   "thing_id_1",
	//  }
	//  err := sdk.Connect(conns, "token")
	//  fmt.Println(err)
	Connect(conns Connection, token string) errors.SDKError

	// Disconnect
	//
	// example:
	//  conns := sdk.Connection{
	//    ChannelID: "channel_id_1",
	//    ThingID:   "thing_id_1",
	//  }
	//  err := sdk.Disconnect(conns, "token")
	//  fmt.Println(err)
	Disconnect(connIDs Connection, token string) errors.SDKError

	// ConnectThing connects thing to specified channel by id.
	//
	// The `ConnectThing` method calls the `CreateThingPolicy` method under the hood.
	//
	// example:
	//  err := sdk.ConnectThing("thingID", "channelID", "token")
	//  fmt.Println(err)
	ConnectThing(thingID, chanID, token string) errors.SDKError

	// DisconnectThing disconnect thing from specified channel by id.
	//
	// The `DisconnectThing` method calls the `DeleteThingPolicy` method under the hood.
	//
	// example:
	//  err := sdk.DisconnectThing("thingID", "channelID", "token")
	//  fmt.Println(err)
	DisconnectThing(thingID, chanID, token string) errors.SDKError

	// SendMessage send message to specified channel.
	//
	// example:
	//  msg := '[{"bn":"some-base-name:","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]'
	//  err := sdk.SendMessage("channelID", msg, "thingSecret")
	//  fmt.Println(err)
	SendMessage(chanID, msg, key string) errors.SDKError

	// ReadMessages read messages of specified channel.
	//
	// example:
	//  msgs, _ := sdk.ReadMessages("channelID", "token")
	//  fmt.Println(msgs)
	ReadMessages(chanID, token string) (MessagesPage, errors.SDKError)

	// SetContentType sets message content type.
	//
	// example:
	//  err := sdk.SetContentType("application/json")
	//  fmt.Println(err)
	SetContentType(ct ContentType) errors.SDKError

	// Health returns service health check.
	//
	// example:
	//  health, _ := sdk.Health("service")
	//  fmt.Println(health)
	Health(service string) (HealthInfo, errors.SDKError)

	// AddBootstrap add bootstrap configuration
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  id, _ := sdk.AddBootstrap(cfg, "token")
	//  fmt.Println(id)
	AddBootstrap(cfg BootstrapConfig, token string) (string, errors.SDKError)

	// View returns Thing Config with given ID belonging to the user identified by the given token.
	//
	// example:
	//  bootstrap, _ := sdk.ViewBootstrap("id", "token")
	//  fmt.Println(bootstrap)
	ViewBootstrap(id, token string) (BootstrapConfig, errors.SDKError)

	// Update updates editable fields of the provided Config.
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  err := sdk.UpdateBootstrap(cfg, "token")
	//  fmt.Println(err)
	UpdateBootstrap(cfg BootstrapConfig, token string) errors.SDKError

	// Update bootstrap config certificates.
	//
	// example:
	//  err := sdk.UpdateBootstrapCerts("id", "clientCert", "clientKey", "ca", "token")
	//  fmt.Println(err)
	UpdateBootstrapCerts(id string, clientCert, clientKey, ca string, token string) (BootstrapConfig, errors.SDKError)

	// UpdateBootstrapConnection updates connections performs update of the channel list corresponding Thing is connected to.
	//
	// example:
	//  err := sdk.UpdateBootstrapConnection("id", []string{"channel1", "channel2"}, "token")
	//  fmt.Println(err)
	UpdateBootstrapConnection(id string, channels []string, token string) errors.SDKError

	// Remove removes Config with specified token that belongs to the user identified by the given token.
	//
	// example:
	//  err := sdk.RemoveBootstrap("id", "token")
	//  fmt.Println(err)
	RemoveBootstrap(id, token string) errors.SDKError

	// Bootstrap returns Config to the Thing with provided external ID using external key.
	//
	// example:
	//  bootstrap, _ := sdk.Bootstrap("externalID", "externalKey")
	//  fmt.Println(bootstrap)
	Bootstrap(externalID, externalKey string) (BootstrapConfig, errors.SDKError)

	// BootstrapSecure retrieves a configuration with given external ID and encrypted external key.
	//
	// example:
	//  bootstrap, _ := sdk.BootstrapSecure("externalID", "externalKey")
	//  fmt.Println(bootstrap)
	BootstrapSecure(externalID, externalKey string) (BootstrapConfig, errors.SDKError)

	// Bootstraps retrieves a list of managed configs.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//  }
	//  bootstraps, _ := sdk.Bootstraps(pm, "token")
	//  fmt.Println(bootstraps)
	Bootstraps(pm PageMetadata, token string) (BootstrapPage, errors.SDKError)

	// Whitelist updates Thing state Config with given ID belonging to the user identified by the given token.
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  err := sdk.Whitelist(cfg, "token")
	//  fmt.Println(err)
	Whitelist(cfg BootstrapConfig, token string) errors.SDKError

	// IssueCert issues a certificate for a thing required for mTLS.
	//
	// example:
	//  cert, _ := sdk.IssueCert("thingID", "valid", "token")
	//  fmt.Println(cert)
	IssueCert(thingID, valid, token string) (Cert, errors.SDKError)

	// ViewCert returns a certificate given certificate ID
	//
	// example:
	//  cert, _ := sdk.ViewCert("certID", "token")
	//  fmt.Println(cert)
	ViewCert(certID, token string) (Cert, errors.SDKError)

	// ViewCertByThing retrieves a list of certificates' serial IDs for a given thing ID.
	//
	// example:
	//  cserial, _ := sdk.ViewCertByThing("thingID", "token")
	//  fmt.Println(cserial)
	ViewCertByThing(thingID, token string) (CertSerials, errors.SDKError)

	// RevokeCert revokes certificate for thing with thingID
	//
	// example:
	//  tm, _ := sdk.RevokeCert("thingID", "token")
	//  fmt.Println(tm)
	RevokeCert(thingID, token string) (time.Time, errors.SDKError)

	// CreateSubscription creates a new subscription
	//
	// example:
	//  subscription, _ := sdk.CreateSubscription("topic", "contact", "token")
	//  fmt.Println(subscription)
	CreateSubscription(topic, contact, token string) (string, errors.SDKError)

	// ListSubscriptions list subscriptions given list parameters.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//  }
	//  subscriptions, _ := sdk.ListSubscriptions(pm, "token")
	//  fmt.Println(subscriptions)
	ListSubscriptions(pm PageMetadata, token string) (SubscriptionPage, errors.SDKError)

	// ViewSubscription retrieves a subscription with the provided id.
	//
	// example:
	//  subscription, _ := sdk.ViewSubscription("id", "token")
	//  fmt.Println(subscription)
	ViewSubscription(id, token string) (Subscription, errors.SDKError)

	// DeleteSubscription removes a subscription with the provided id.
	//
	// example:
	//  err := sdk.DeleteSubscription("id", "token")
	//  fmt.Println(err)
	DeleteSubscription(id, token string) errors.SDKError
}

type mfSDK struct {
	bootstrapURL   string
	certsURL       string
	httpAdapterURL string
	readerURL      string
	thingsURL      string
	usersURL       string
	HostURL        string

	msgContentType ContentType
	client         *http.Client
}

// Config contains sdk configuration parameters.
type Config struct {
	BootstrapURL   string
	CertsURL       string
	HTTPAdapterURL string
	ReaderURL      string
	ThingsURL      string
	UsersURL       string
	HostURL        string

	MsgContentType  ContentType
	TLSVerification bool
}

// NewSDK returns new mainflux SDK instance.
func NewSDK(conf Config) SDK {
	return &mfSDK{
		bootstrapURL:   conf.BootstrapURL,
		certsURL:       conf.CertsURL,
		httpAdapterURL: conf.HTTPAdapterURL,
		readerURL:      conf.ReaderURL,
		thingsURL:      conf.ThingsURL,
		usersURL:       conf.UsersURL,
		HostURL:        conf.HostURL,

		msgContentType: conf.MsgContentType,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !conf.TLSVerification,
				},
			},
		},
	}
}

// processRequest creates and send a new HTTP request, and checks for errors in the HTTP response.
// It then returns the response headers, the response body, and the associated error(s) (if any).
func (sdk mfSDK) processRequest(method, url, token string, data []byte, headers map[string]string, expectedRespCodes ...int) (http.Header, []byte, errors.SDKError) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	// Sets a default value for the Content-Type.
	// Overridden if Content-Type is passed in the headers arguments.
	req.Header.Add("Content-Type", string(CTJSON))

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if token != "" {
		if !strings.Contains(token, ThingPrefix) {
			token = BearerPrefix + token
		}
		req.Header.Set("Authorization", token)
	}

	resp, err := sdk.client.Do(req)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}
	defer resp.Body.Close()

	sdkerr := errors.CheckError(resp, expectedRespCodes...)
	if sdkerr != nil {
		return make(http.Header), []byte{}, sdkerr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	return resp.Header, body, nil
}

func (sdk mfSDK) withQueryParams(baseURL, endpoint string, pm PageMetadata) (string, error) {
	q, err := pm.query()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s?%s", baseURL, endpoint, q), nil
}

func (pm PageMetadata) query() (string, error) {
	q := url.Values{}
	if pm.Offset != 0 {
		q.Add("offset", strconv.FormatUint(pm.Offset, 10))
	}
	if pm.Limit != 0 {
		q.Add("limit", strconv.FormatUint(pm.Limit, 10))
	}
	if pm.Total != 0 {
		q.Add("total", strconv.FormatUint(pm.Total, 10))
	}
	if pm.Level != 0 {
		q.Add("level", strconv.FormatUint(pm.Level, 10))
	}
	if pm.Email != "" {
		q.Add("email", pm.Email)
	}
	if pm.Name != "" {
		q.Add("name", pm.Name)
	}
	if pm.Type != "" {
		q.Add("type", pm.Type)
	}
	if pm.Visibility != "" {
		q.Add("visibility", pm.Visibility)
	}
	if pm.Status != "" {
		q.Add("status", pm.Status)
	}
	if pm.Metadata != nil {
		md, err := json.Marshal(pm.Metadata)
		if err != nil {
			return "", errors.NewSDKError(err)
		}
		q.Add("metadata", string(md))
	}
	if pm.Action != "" {
		q.Add("action", pm.Action)
	}
	if pm.Subject != "" {
		q.Add("subject", pm.Subject)
	}
	if pm.Object != "" {
		q.Add("object", pm.Object)
	}
	if pm.Tag != "" {
		q.Add("tag", pm.Tag)
	}
	if pm.Owner != "" {
		q.Add("owner", pm.Owner)
	}
	if pm.SharedBy != "" {
		q.Add("shared_by", pm.SharedBy)
	}
	if pm.Topic != "" {
		q.Add("topic", pm.Topic)
	}
	if pm.Contact != "" {
		q.Add("contact", pm.Contact)
	}
	if pm.State != "" {
		q.Add("state", pm.State)
	}

	return q.Encode(), nil
}
