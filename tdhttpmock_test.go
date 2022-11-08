// Copyright (c) 2022, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tdhttpmock_test

import (
	"io/ioutil" //nolint: staticcheck
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"

	"github.com/maxatome/tdhttpmock"
)

func TestBodyHeader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

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

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body("STRING").WithName("50-body-STRING"),
		httpmock.NewStringResponder(200, "OK-50"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body([]byte("BYTES")).WithName("40-body-BYTES"),
		httpmock.NewStringResponder(200, "OK-40"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body(td.All(td.Contains("FOO"), td.Contains("BAR"))).
			WithName("60-body-FOO-BAR"),
		httpmock.NewStringResponder(200, "OK-60"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body(td.Empty()).WithName("70-body-empty"),
		httpmock.NewStringResponder(200, "OK-70"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.Body(666).WithName("00-never-match"),
		httpmock.NewStringResponder(200, "BAD-00"),
	)

	assert := td.Assert(t)

	assert.RunAssertRequire("20-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("42 test"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-20")))
	})

	assert.RunAssertRequire("not found", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("x test"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("Not found")))
	})

	assert.RunAssertRequire("10-body+header", func(assert, require *td.T) {
		req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader("42 test"))
		require.CmpNoError(err)

		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("X-Custom", "YES")

		resp, err := http.DefaultClient.Do(req)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-10")))
	})

	assert.RunAssertRequire("40-body-BYTES", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("BYTES"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-40")))
	})

	assert.RunAssertRequire("50-body-STRING", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("STRING"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-50")))
	})

	assert.RunAssertRequire("60-body-FOO-BAR", func(assert, require *td.T) {
		resp, err := http.Post("/test", "text/plain", strings.NewReader("--FOO--BAR--"))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-60")))
	})

	assert.RunAssertRequire("70-body-empty", func(assert, require *td.T) {
		resp, err := http.Post("/test", "", nil)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-70")))
	})
}
