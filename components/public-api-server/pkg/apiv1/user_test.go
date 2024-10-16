// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package apiv1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	v1 "github.com/gitpod-io/gitpod/components/public-api/go/experimental/v1"
	"github.com/gitpod-io/gitpod/components/public-api/go/experimental/v1/v1connect"
	protocol "github.com/gitpod-io/gitpod/gitpod-protocol"
	"github.com/gitpod-io/gitpod/public-api-server/pkg/auth"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetAuthenticatedUser(t *testing.T) {
	t.Run("proxies request to server", func(t *testing.T) {
		serverMock, client := setupUserService(t)

		user := newUser(&protocol.User{
			Name: "John",
		})

		serverMock.EXPECT().GetLoggedInUser(gomock.Any()).Return(user, nil)

		retrieved, err := client.GetAuthenticatedUser(context.Background(), connect.NewRequest(&v1.GetAuthenticatedUserRequest{}))
		require.NoError(t, err)
		requireEqualProto(t, &v1.GetAuthenticatedUserResponse{
			User: userToAPIResponse(user),
		}, retrieved.Msg)
	})
}

func TestUserService_ListSSHKeys(t *testing.T) {
	t.Run("proxies request to server", func(t *testing.T) {
		serverMock, client := setupUserService(t)

		var keys []*protocol.UserSSHPublicKeyValue
		keys = append(keys, newSSHKey(&protocol.UserSSHPublicKeyValue{
			Name: "test key",
		}))

		serverMock.EXPECT().GetSSHPublicKeys(gomock.Any()).Return(keys, nil)

		var expected []*v1.SSHKey
		for _, k := range keys {
			expected = append(expected, sshKeyToAPIResponse(k))
		}

		retrieved, err := client.ListSSHKeys(context.Background(), connect.NewRequest(&v1.ListSSHKeysRequest{}))
		require.NoError(t, err)
		requireEqualProto(t, &v1.ListSSHKeysResponse{
			Keys: expected,
		}, retrieved.Msg)
	})
}

func setupUserService(t *testing.T) (*protocol.MockAPIInterface, v1connect.UserServiceClient) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	serverMock := protocol.NewMockAPIInterface(ctrl)

	svc := NewUserService(&FakeServerConnPool{
		api: serverMock,
	})

	_, handler := v1connect.NewUserServiceHandler(svc, connect.WithInterceptors(auth.NewServerInterceptor()))

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client := v1connect.NewUserServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(
		auth.NewClientInterceptor("auth-token"),
	))

	return serverMock, client
}

func newUser(t *protocol.User) *protocol.User {
	result := &protocol.User{
		ID:           uuid.New().String(),
		Name:         "John",
		AvatarURL:    "https://avatars.yolo/first.png",
		CreationDate: "2022-10-10T10:10:10.000Z",
	}

	if t.ID != "" {
		result.ID = t.ID
	}

	if t.Name != "" {
		result.Name = t.Name
	}

	if t.CreationDate != "" {
		result.CreationDate = t.CreationDate
	}

	return result
}

func newSSHKey(t *protocol.UserSSHPublicKeyValue) *protocol.UserSSHPublicKeyValue {
	result := &protocol.UserSSHPublicKeyValue{
		ID:           uuid.New().String(),
		Name:         "John",
		Key:          "ssh-ed25519 AAAAB3NzaC1yc2EAAAADAQABAAACAQDCnrN9UdK1bNGPmZfenTW",
		Fingerprint:  "ykjP/b5aqoa3envmXzWpPMCGgEFMu3QvubfSTNrJCMA=",
		CreationTime: "2022-10-10T10:10:10.000Z",
		LastUsedTime: "2022-10-10T10:10:10.000Z",
	}

	if t.ID != "" {
		result.ID = t.ID
	}

	if t.Name != "" {
		result.Name = t.Name
	}

	if t.Key != "" {
		result.Key = t.Key
	}

	if t.CreationTime != "" {
		result.CreationTime = t.CreationTime
	}

	return result
}
