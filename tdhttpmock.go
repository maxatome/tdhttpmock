// Copyright (c) 2022, Maxime Soulé
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package tdhttpmock

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil" //nolint: staticcheck
	"net/http"
	"reflect"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
)

var interfaceType = reflect.TypeOf((*any)(nil)).Elem()

func marshaledBody(
	acceptEmptyBody bool,
	unmarshal func([]byte, any) error,
	expectedBody any,
) httpmock.MatcherFunc {
	return func(req *http.Request) bool {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false
		}

		if !acceptEmptyBody && len(body) == 0 {
			return false
		}

		var bodyType reflect.Type

		// If expectedBody is a TestDeep operator, try to ask it the type
		// behind it.
		if op, ok := expectedBody.(td.TestDeep); ok {
			bodyType = op.TypeBehind()
		} else {
			bodyType = reflect.TypeOf(expectedBody)
		}

		// As the expected body type cannot be guessed, try to
		// unmarshal in an any
		if bodyType == nil {
			bodyType = interfaceType
		}

		bodyPtr := reflect.New(bodyType)

		if unmarshal(body, bodyPtr.Interface()) != nil {
			return false
		}

		return td.EqDeeply(bodyPtr.Elem().Interface(), expectedBody)
	}
}

// Body returns an [httpmock.Matcher] matching request body against
// expectedBody. expectedBody can be a []byte, a string or a
// [td.TestDeep] operator.
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Body("OK!\n"),
//	  httpmock.NewStringResponder(200, "OK"))
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Body(td.Re(`\d+ test`)),
//	  httpmock.NewStringResponder(200, "OK test"))
//
// The name of the returned [httpmock.Matcher] is auto-generated (see
// [httpmock.NewMatcher]). To name it explicitly, use
// [httpmock.Matcher.WithName] as in:
//
//	tdhttpmock.Body("OK!\n").WithName("01-body-OK")
func Body(expectedBody any) httpmock.Matcher {
	return httpmock.NewMatcher("",
		marshaledBody(true,
			func(body []byte, target any) error {
				switch target := target.(type) {
				case *string:
					*target = string(body)
				case *[]byte:
					*target = body
				case *any:
					*target = body
				default:
					// marshaledBody always calls us with target as a pointer
					return fmt.Errorf(
						"Body only accepts expectedBody be a []byte, a string or a TestDeep operator allowing to match these types, but not type %s",
						reflect.TypeOf(target).Elem())
				}
				return nil
			},
			expectedBody))
}

// JSONBody returns an [httpmock.Matcher] expecting a JSON request body
// that can be [json.Unmarshal]'ed and that matches expectedBody.
// expectedBody can be any type one can [json.Unmarshal] into, or a
// [td.TestDeep] operator.
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.JSONBody(Person{
//	    ID:   42,
//	    Name: "Bob",
//	    Age:  26,
//	  }),
//	  httpmock.NewStringResponder(200, "OK bob"))
//
// The same using [td.JSON]:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.JSONBody(td.JSON(`
//	    {
//	      "id":   NotZero(),
//	      "name": "Bob",
//	      "age":  26
//	    }`)),
//	  httpmock.NewStringResponder(200, "OK bob"))
//
// Note also the existence of [td.JSONPointer]:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.JSONBody(td.JSONPointer("/name", "Bob")),
//	  httpmock.NewStringResponder(200, "OK bob"))
//
// The name of the returned [httpmock.Matcher] is auto-generated (see
// [httpmock.NewMatcher]). To name it explicitly, use
// [httpmock.Matcher.WithName] as in:
//
//	tdhttpmock.JSONBody(td.JSONPointer("/name", "Bob")).WithName("01-bob")
func JSONBody(expectedBody any) httpmock.Matcher {
	return httpmock.NewMatcher("",
		marshaledBody(false, json.Unmarshal, expectedBody))
}

// XMLBody returns an [httpmock.Matcher] expecting an XML request
// body that can be [xml.Unmarshal]'ed and that matches
// expectedBody. expectedBody can be any type one can [xml.Unmarshal]
// into, or a [td.TestDeep] operator.
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.XMLBody(Person{
//	    ID:   42,
//	    Name: "Bob",
//	    Age:  26,
//	  }),
//	  httpmock.NewStringResponder(200, "OK bob"))
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.XMLBody(td.SStruct(
//	    Person{
//	      Name: "Bob",
//	      Age:  26,
//	    },
//	    td.StructFields{
//	      "ID": td.NotZero(),
//	    })),
//	  httpmock.NewStringResponder(200, "OK bob"))
//
// The name of the returned [httpmock.Matcher] is auto-generated (see
// [httpmock.NewMatcher]). To name it explicitly, use
// [httpmock.Matcher.WithName] as in:
//
//	tdhttpmock.XMLBody(td.Struct(Person{Name: "Bob"})).WithName("01-bob")
func XMLBody(expectedBody any) httpmock.Matcher {
	return httpmock.NewMatcher("",
		marshaledBody(false, xml.Unmarshal, expectedBody))
}

// Header returns an [httpmock.Matcher] matching request header against
// expectedHeader. expectedHeader can be a [http.Header] or a
// [td.TestDeep] operator. Keep in mind that if it is a [http.Header],
// it has to match exactly the response header. Often only the
// presence of a header key is needed:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Header(td.ContainsKey("X-Custom")),
//	  httpmock.NewStringResponder(200, "OK custom"))
//
// or some specific key, value pairs:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Header(td.SuperMapOf(
//	    http.Header{
//	    "X-Account": []string{"Bob"},
//	    },
//	    td.MapEntries{
//	      "X-Token": td.Bag(td.Re(`^[a-z0-9-]{32}\z`)),
//	    },
//	  )),
//	  httpmock.NewStringResponder(200, "OK account"))
//
// The name of the returned [httpmock.Matcher] is auto-generated (see
// [httpmock.NewMatcher]). To name it explicitly, use
// [httpmock.Matcher.WithName] as in:
//
//	tdhttpmock.Header(td.ContainsKey("X-Custom")).WithName("01-header-custom")
func Header(expectedHader any) httpmock.Matcher {
	return httpmock.NewMatcher("",
		func(req *http.Request) bool {
			return td.EqDeeply(req.Header, td.Lax(expectedHader))
		})
}

// Cookies returns an [httpmock.Matcher] matching request cookies
// against expectedCookies. expectedCookies can be a [][*http.Cookie]
// or a [td.TestDeep] operator. Keep in mind that if it is a
// [][*http.Cookie], it has to match exactly the response
// cookies. Often only the presence of a cookie key is needed:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Cookies(td.SuperBagOf(td.Smuggle("Name", "cookie_session"))),
//	  httpmock.NewStringResponder(200, "OK session"))
//
// To make tests easier, [http.Cookie.Raw] and [http.Cookie.RawExpires] fields
// of each [*http.Cookie] are zeroed before doing the comparison. So no need
// to fill them when comparing against a simple literal as in:
//
//	httpmock.RegisterMatcherResponder(
//	  http.MethodPost,
//	  "/test",
//	  tdhttpmock.Cookies([]*http.Cookies{
//	    {Name: "cookieName1", Value: "cookieValue1"},
//	    {Name: "cookieName2", Value: "cookieValue2"},
//	  }),
//	  httpmock.NewStringResponder(200, "OK cookies"))
//
// The name of the returned [httpmock.Matcher] is auto-generated (see
// [httpmock.NewMatcher]). To name it explicitly, use
// [httpmock.Matcher.WithName] as in:
//
//	tdhttpmock.Cookies([]*http.Cookies{}).WithName("01-cookies")
func Cookies(expectedCookies any) httpmock.Matcher {
	return httpmock.NewMatcher("",
		func(req *http.Request) bool {
			// Empty Raw* fields to make comparisons easier
			cookies := req.Cookies()
			for _, c := range cookies {
				c.RawExpires, c.Raw = "", ""
			}
			return td.EqDeeply(cookies, td.Lax(expectedCookies))
		})
}
