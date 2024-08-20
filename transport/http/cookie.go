// Package http
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package http

import (
	"github.com/valyala/fasthttp"
)

// CookieOption cookie option
type CookieOption func(cookie *fasthttp.Cookie)

// CookieSameSite cookie same site
type CookieSameSite = fasthttp.CookieSameSite

const (
	// CookieSameSiteDisabled removes the SameSite flag.
	CookieSameSiteDisabled CookieSameSite = iota
	// CookieSameSiteDefaultMode sets the SameSite flag.
	CookieSameSiteDefaultMode
	// CookieSameSiteLaxMode sets the SameSite flag with the "Lax" parameter.
	CookieSameSiteLaxMode
	// CookieSameSiteStrictMode sets the SameSite flag with the "Strict" parameter.
	CookieSameSiteStrictMode
	// CookieSameSiteNoneMode sets the SameSite flag with the "None" parameter.
	// See https://tools.ietf.org/html/draft-west-cookie-incrementalism-00
	CookieSameSiteNoneMode // third-party cookies are phasing out, use Partitioned cookies instead
)

// CookieWithMaxAge cookie max age
func CookieWithMaxAge(maxAge int) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetMaxAge(maxAge)
	}
}

// CookieWithPath cookie path
func CookieWithPath(path string) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetPath(path)
	}
}

// CookieWithDomain cookie domain
func CookieWithDomain(domain string) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetDomain(domain)
	}
}

// CookieWithSecure cookie secure
func CookieWithSecure(secure bool) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetSecure(secure)
	}
}

// CookieWithHTTPOnly cookie http only
func CookieWithHTTPOnly(httpOnly bool) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetHTTPOnly(httpOnly)
	}
}

// SetSameSite cookie
func SetSameSite(sameSite CookieSameSite) CookieOption {
	return func(cookie *fasthttp.Cookie) {
		cookie.SetSameSite(sameSite)
	}
}
