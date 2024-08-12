// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
)

func handleAzureRequestError(err error, actionTmpl string, parts ...interface{}) (bool, error) {
	action := fmt.Sprintf(actionTmpl, parts...)

	var respError *azcore.ResponseError
	if !errors.As(err, &respError) {
		// Error should be a azcore.ResponseError otherwise something
		// bad has happened in the client.
		return false, provider.FatalErrorf("unexpected error from azure while %s: %w", action, err)
	}

	sc := respError.StatusCode
	switch {
	case sc >= 400 && sc < 500:
		// Client errors (BadRequest/Unauthorized etc) are Fatal. We
		// also return true to indicate we have NotFound which is a
		// special case in a lot of processing.
		return sc == http.StatusNotFound, provider.FatalErrorf("error from azure while %s: %w", action, err)
	default:
		// Everything else is a normal error which can be
		// logged as a failure and then the reconciler will try
		// again on the next loop.
		return false, fmt.Errorf("error from azure while %s: %w", action, err)
	}
}

func ensureDeleted(resourceType string, getFunc func() error, deleteFunc func() error, estimateTime time.Duration) error {
	err := getFunc()
	if err != nil {
		notFound, err := handleAzureRequestError(err, "getting %s", resourceType)
		// NotFound means that the resource has been deleted
		// successfully, all other errors are raised.
		if notFound {
			return nil
		}
		return err
	}

	err = deleteFunc()
	if err != nil {
		_, err := handleAzureRequestError(err, "deleting %s", resourceType)
		return err
	}

	return provider.RetryableErrorf(estimateTime, "%s delete issued", resourceType)
}
