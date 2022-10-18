// Copyright (c) 2022, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tdhttpmock_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"

	"github.com/maxatome/tdhttpmock"
)

func TestTdhttpmock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, "Not found"))

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body(td.Re(`\d+ test`)).WithName("20-body"),
		httpmock.NewStringResponder(200, "OK-20"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body(td.Re(`\d+ test`)).
			And(tdhttpmock.Header(td.ContainsKey("X-Custom"))).
			WithName("10-body+header"),
		httpmock.NewStringResponder(200, "OK-10"),
	)

	assert := td.Assert(t)

	assert.RunAssertRequire("20-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("42 test"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(io.ReadAll, td.String("OK-20")))
	})

	assert.RunAssertRequire("not found", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("x test"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
		assert.Cmp(resp.Body, td.Smuggle(io.ReadAll, td.String("Not found")))
	})

	assert.RunAssertRequire("10-body+header", func(assert, require *td.T) {
		req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader("42 test"))
		require.CmpNoError(err)

		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("X-Custom", "YES")

		resp, err := http.DefaultClient.Do(req)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(io.ReadAll, td.String("OK-10")))
	})
}
