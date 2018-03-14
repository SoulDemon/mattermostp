// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sqlstore

import (
	"testing"

	"github.com/SoulDemon/mattermostp/store/storetest"
)

func TestUserAccessTokenStore(t *testing.T) {
	StoreTest(t, storetest.TestUserAccessTokenStore)
}
