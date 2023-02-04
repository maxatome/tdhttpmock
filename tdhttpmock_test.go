// Copyright (c) 2022-2023, Maxime Soulé
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

func TestMatcherName(t *testing.T) {
	td.Cmp(t, tdhttpmock.Body(td.Empty()).Name(), td.Re(`/tdhttpmock_test.go:\d+$`))
}

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

func TestJSONBody(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, "Not found"))

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{"name":"bob"}`)).WithName("20-body"),
		httpmock.NewStringResponder(200, "OK-20"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{"age":23}`)).WithName("10-body"),
		httpmock.NewStringResponder(200, "OK-10"),
	)

	assert := td.Assert(t)

	assert.RunAssertRequire("20-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/json",
			strings.NewReader(`{"name":"bob","age":66}`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-20")))
	})

	assert.RunAssertRequire("10-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/json",
			strings.NewReader(`{"name":"bob","age":23}`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-10")))
	})

	assert.RunAssertRequire("no match", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/json",
			strings.NewReader(`{"name":"alice","age":32}`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
	})

	assert.RunAssertRequire("empty", func(assert, require *td.T) {
		resp, err := http.Post("/test", "", nil)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
	})
}

func TestXMLBody(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type XBody struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, "Not found"))

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.XMLBody(td.Struct(XBody{Name: "bob"}, nil)).WithName("20-body"),
		httpmock.NewStringResponder(200, "OK-20"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"/test",
		tdhttpmock.XMLBody(td.Struct(XBody{Age: 23}, nil)).WithName("10-body"),
		httpmock.NewStringResponder(200, "OK-10"),
	)

	assert := td.Assert(t)

	assert.RunAssertRequire("20-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/xml",
			strings.NewReader(`<XBody><name>bob</name><age>66</age></XBody>`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-20")))
	})

	assert.RunAssertRequire("10-body", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/xml",
			strings.NewReader(`<XBody><name>bob</name><age>23</age></XBody>`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-10")))
	})

	assert.RunAssertRequire("no match", func(assert, require *td.T) {
		resp, err := http.Post("/test", "application/xml",
			strings.NewReader(`<XBody><name>alice</name><age>32</age></XBody>`))
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
	})

	assert.RunAssertRequire("empty", func(assert, require *td.T) {
		resp, err := http.Post("/test", "", nil)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
	})
}

func TestCookies(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, "Not found"))

	httpmock.RegisterMatcherResponder(
		http.MethodGet,
		"/test",
		tdhttpmock.Cookies([]*http.Cookie{
			{Name: "first", Value: "cookie1"},
			{Name: "second", Value: "cookie2"},
		}).WithName("20-cookies"),
		httpmock.NewStringResponder(200, "OK-20"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodGet,
		"/test",
		tdhttpmock.Cookies(td.Bag(
			&http.Cookie{Name: "first", Value: "cookie1"},
			&http.Cookie{Name: "third", Value: "cookie3"},
		)).WithName("10-cookies"),
		httpmock.NewStringResponder(200, "OK-10"),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodGet,
		"/test",
		tdhttpmock.Cookies(td.SuperBagOf(
			&http.Cookie{Name: "third", Value: "cookie3"},
		)).WithName("30-cookies"),
		httpmock.NewStringResponder(200, "OK-30"),
	)

	setCookie := func(req *http.Request, cookie *http.Cookie) {
		if v := cookie.String(); v != "" {
			req.Header.Add("Cookie", v)
		}
	}

	assert := td.Assert(t)

	assert.RunAssertRequire("20-cookies", func(assert, require *td.T) {
		req, err := http.NewRequest("GET", "/test", nil)
		require.CmpNoError(err)

		setCookie(req, &http.Cookie{Name: "first", Value: "cookie1"})
		setCookie(req, &http.Cookie{Name: "second", Value: "cookie2"})

		resp, err := http.DefaultClient.Do(req)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-20")))
	})

	assert.RunAssertRequire("10-cookies", func(assert, require *td.T) {
		req, err := http.NewRequest("GET", "/test", nil)
		require.CmpNoError(err)

		setCookie(req, &http.Cookie{Name: "third", Value: "cookie3"})
		setCookie(req, &http.Cookie{Name: "first", Value: "cookie1"})

		resp, err := http.DefaultClient.Do(req)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-10")))
	})

	assert.RunAssertRequire("30-cookies", func(assert, require *td.T) {
		req, err := http.NewRequest("GET", "/test", nil)
		require.CmpNoError(err)

		setCookie(req, &http.Cookie{Name: "third", Value: "cookie3"})
		setCookie(req, &http.Cookie{Name: "another", Value: "cookieX"})

		resp, err := http.DefaultClient.Do(req)
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-30")))
	})

	assert.RunAssertRequire("no cookies", func(assert, require *td.T) {
		resp, err := http.Get("/test")
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 404)
	})

	httpmock.RegisterMatcherResponder(
		http.MethodGet,
		"/test",
		tdhttpmock.Cookies(td.Empty()).WithName("NO-cookies"),
		httpmock.NewStringResponder(200, "OK-NO"),
	)

	assert.RunAssertRequire("catch no cookies", func(assert, require *td.T) {
		resp, err := http.Get("/test")
		require.CmpNoError(err)
		assert.Cmp(resp.StatusCode, 200)
		assert.Cmp(resp.Body, td.Smuggle(ioutil.ReadAll, td.String("OK-NO")))
	})
}
