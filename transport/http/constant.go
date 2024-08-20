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
	"math"
)

const (
	AbortIndex int8 = math.MaxInt8 / 2 // AbortIndex abort page index
)

type (
	userContextKey struct{}
	httpCtxKey     struct{}
)

const (
	StatusContinue           = 100 // StatusContinue RFC 9110, 15.2.1
	StatusSwitchingProtocols = 101 // StatusSwitchingProtocols RFC 9110, 15.2.2
	StatusProcessing         = 102 // StatusProcessing RFC 2518, 10.1
	StatusEarlyHints         = 103 // StatusEarlyHints RFC 8297

	StatusOK                          = 200 // StatusOK RFC 9110, 15.3.1
	StatusCreated                     = 201 // StatusCreated RFC 9110, 15.3.2
	StatusAccepted                    = 202 // StatusAccepted RFC 9110, 15.3.3
	StatusNonAuthoritativeInformation = 203 // StatusNonAuthoritativeInformation RFC 9110, 15.3.4
	StatusNoContent                   = 204 // StatusNoContent RFC 9110, 15.3.5
	StatusResetContent                = 205 // StatusResetContent RFC 9110, 15.3.6
	StatusPartialContent              = 206 // StatusPartialContent RFC 9110, 15.3.7
	StatusMultiStatus                 = 207 // StatusMultiStatus RFC 4918, 11.1
	StatusAlreadyReported             = 208 // StatusAlreadyReported RFC 5842, 7.1
	StatusIMUsed                      = 226 // StatusIMUsed RFC 3229, 10.4.1

	StatusMultipleChoices   = 300 // StatusMultipleChoices RFC 9110, 15.4.1
	StatusMovedPermanently  = 301 // StatusMovedPermanently RFC 9110, 15.4.2
	StatusFound             = 302 // StatusFound RFC 9110, 15.4.3
	StatusSeeOther          = 303 // StatusSeeOther RFC 9110, 15.4.4
	StatusNotModified       = 304 // StatusNotModified RFC 9110, 15.4.5
	StatusUseProxy          = 305 // StatusUseProxy RFC 9110, 15.4.6
	StatusSwitchProxy       = 306 // StatusSwitchProxy RFC 9110, 15.4.7 (Unused)
	StatusTemporaryRedirect = 307 // StatusTemporaryRedirect RFC 9110, 15.4.8
	StatusPermanentRedirect = 308 // StatusPermanentRedirect RFC 9110, 15.4.9

	StatusBadRequest                   = 400 // StatusBadRequest RFC 9110, 15.5.1
	StatusUnauthorized                 = 401 // StatusUnauthorized RFC 9110, 15.5.2
	StatusPaymentRequired              = 402 // StatusPaymentRequired RFC 9110, 15.5.3
	StatusForbidden                    = 403 // StatusForbidden RFC 9110, 15.5.4
	StatusNotFound                     = 404 // StatusNotFound RFC 9110, 15.5.5
	StatusMethodNotAllowed             = 405 // StatusMethodNotAllowed RFC 9110, 15.5.6
	StatusNotAcceptable                = 406 // StatusNotAcceptable RFC 9110, 15.5.7
	StatusProxyAuthRequired            = 407 // StatusProxyAuthRequired RFC 9110, 15.5.8
	StatusRequestTimeout               = 408 // StatusRequestTimeout RFC 9110, 15.5.9
	StatusConflict                     = 409 // StatusConflict RFC 9110, 15.5.10
	StatusGone                         = 410 // StatusGone RFC 9110, 15.5.11
	StatusLengthRequired               = 411 // StatusLengthRequired RFC 9110, 15.5.12
	StatusPreconditionFailed           = 412 // StatusPreconditionFailed RFC 9110, 15.5.13
	StatusRequestEntityTooLarge        = 413 // StatusRequestEntityTooLarge RFC 9110, 15.5.14
	StatusRequestURITooLong            = 414 // StatusRequestURITooLong RFC 9110, 15.5.15
	StatusUnsupportedMediaType         = 415 // StatusUnsupportedMediaType RFC 9110, 15.5.16
	StatusRequestedRangeNotSatisfiable = 416 // StatusRequestedRangeNotSatisfiable RFC 9110, 15.5.17
	StatusExpectationFailed            = 417 // StatusExpectationFailed RFC 9110, 15.5.18
	StatusTeapot                       = 418 // StatusTeapot RFC 9110, 15.5.19 (Unused)
	StatusMisdirectedRequest           = 421 // StatusMisdirectedRequest RFC 9110, 15.5.20
	StatusUnprocessableEntity          = 422 // StatusUnprocessableEntity RFC 9110, 15.5.21
	StatusLocked                       = 423 // StatusLocked RFC 4918, 11.3
	StatusFailedDependency             = 424 // StatusFailedDependency RFC 4918, 11.4
	StatusTooEarly                     = 425 // StatusTooEarly RFC 8470, 5.2.
	StatusUpgradeRequired              = 426 // StatusUpgradeRequired RFC 9110, 15.5.22
	StatusPreconditionRequired         = 428 // StatusPreconditionRequired RFC 6585, 3
	StatusTooManyRequests              = 429 // StatusTooManyRequests RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // StatusRequestHeaderFieldsTooLarge RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // StatusUnavailableForLegalReasons RFC 7725, 3

	StatusInternalServerError           = 500 // StatusInternalServerError RFC 9110, 15.6.1
	StatusNotImplemented                = 501 // StatusNotImplemented RFC 9110, 15.6.2
	StatusBadGateway                    = 502 // StatusBadGateway RFC 9110, 15.6.3
	StatusServiceUnavailable            = 503 // StatusServiceUnavailable RFC 9110, 15.6.4
	StatusGatewayTimeout                = 504 // StatusGatewayTimeout RFC 9110, 15.6.5
	StatusHTTPVersionNotSupported       = 505 // StatusHTTPVersionNotSupported RFC 9110, 15.6.6
	StatusVariantAlsoNegotiates         = 506 // StatusVariantAlsoNegotiates RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // StatusInsufficientStorage RFC 4918, 11.5
	StatusLoopDetected                  = 508 // StatusLoopDetected RFC 5842, 7.2
	StatusNotExtended                   = 510 // StatusNotExtended RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // StatusNetworkAuthenticationRequired RFC 6585, 6
)

const (
	default404Body = "404 page not found"
	default405Body = "405 method not allowed"
)

// http method
const (
	MethodGet     = "GET"     // RFC 7231, 4.3.1
	MethodHead    = "HEAD"    // RFC 7231, 4.3.2
	MethodPost    = "POST"    // RFC 7231, 4.3.3
	MethodPut     = "PUT"     // RFC 7231, 4.3.4
	MethodPatch   = "PATCH"   // RFC 5789
	MethodDelete  = "DELETE"  // RFC 7231, 4.3.5
	MethodConnect = "CONNECT" // RFC 7231, 4.3.6
	MethodOptions = "OPTIONS" // RFC 7231, 4.3.7
	MethodTrace   = "TRACE"   // RFC 7231, 4.3.8
	MethodAny     = "*"
)

// statusMessage 状态码对应的默认消息
var statusMessage = []string{
	100: "Continue",            // StatusContinue
	101: "Switching Protocols", // StatusSwitchingProtocols
	102: "Processing",          // StatusProcessing
	103: "Early Hints",         // StatusEarlyHints

	200: "OK",                            // StatusOK
	201: "Created",                       // StatusCreated
	202: "Accepted",                      // StatusAccepted
	203: "Non-Authoritative Information", // StatusNonAuthoritativeInformation
	204: "No Content",                    // StatusNoContent
	205: "Reset Content",                 // StatusResetContent
	206: "Partial Content",               // StatusPartialContent
	207: "Multi-Status",                  // StatusMultiStatus
	208: "Already Reported",              // StatusAlreadyReported
	226: "IM Used",                       // StatusIMUsed

	300: "Multiple Choices",   // StatusMultipleChoices
	301: "Moved Permanently",  // StatusMovedPermanently
	302: "Found",              // StatusFound
	303: "See Other",          // StatusSeeOther
	304: "Not Modified",       // StatusNotModified
	305: "Use Proxy",          // StatusUseProxy
	306: "Switch Proxy",       // StatusSwitchProxy
	307: "Temporary Redirect", // StatusTemporaryRedirect
	308: "Permanent Redirect", // StatusPermanentRedirect

	400: "Bad Request",                     // StatusBadRequest
	401: "Unauthorized",                    // StatusUnauthorized
	402: "Payment Required",                // StatusPaymentRequired
	403: "Forbidden",                       // StatusForbidden
	404: "Not Found",                       // StatusNotFound
	405: "Method Not Allowed",              // StatusMethodNotAllowed
	406: "Not Acceptable",                  // StatusNotAcceptable
	407: "Proxy Authentication Required",   // StatusProxyAuthRequired
	408: "Request Timeout",                 // StatusRequestTimeout
	409: "Conflict",                        // StatusConflict
	410: "Gone",                            // StatusGone
	411: "Length Required",                 // StatusLengthRequired
	412: "Precondition Failed",             // StatusPreconditionFailed
	413: "Request Entity Too Large",        // StatusRequestEntityTooLarge
	414: "Request URI Too Long",            // StatusRequestURITooLong
	415: "Unsupported Media Type",          // StatusUnsupportedMediaType
	416: "Requested Range Not Satisfiable", // StatusRequestedRangeNotSatisfiable
	417: "Expectation Failed",              // StatusExpectationFailed
	418: "I'm a teapot",                    // StatusTeapot
	421: "Misdirected Request",             // StatusMisdirectedRequest
	422: "Unprocessable Entity",            // StatusUnprocessableEntity
	423: "Locked",                          // StatusLocked
	424: "Failed Dependency",               // StatusFailedDependency
	425: "Too Early",                       // StatusTooEarly
	426: "Upgrade Required",                // StatusUpgradeRequired
	428: "Precondition Required",           // StatusPreconditionRequired
	429: "Too Many Requests",               // StatusTooManyRequests
	431: "Request Header Fields Too Large", // StatusRequestHeaderFieldsTooLarge
	451: "Unavailable For Legal Reasons",   // StatusUnavailableForLegalReasons

	500: "Internal Server Error",           // StatusInternalServerError
	501: "Not Implemented",                 // StatusNotImplemented
	502: "Bad Gateway",                     // StatusBadGateway
	503: "Service Unavailable",             // StatusServiceUnavailable
	504: "Gateway Timeout",                 // StatusGatewayTimeout
	505: "HTTP Version Not Supported",      // StatusHTTPVersionNotSupported
	506: "Variant Also Negotiates",         // StatusVariantAlsoNegotiates
	507: "Insufficient Storage",            // StatusInsufficientStorage
	508: "Loop Detected",                   // StatusLoopDetected
	510: "Not Extended",                    // StatusNotExtended
	511: "Network Authentication Required", // StatusNetworkAuthenticationRequired
}

// MIME types that are commonly used
const (
	MIMETextXML         = "text/xml"
	MIMETextHTML        = "text/html"
	MIMETextPlain       = "text/plain"
	MIMETextJavaScript  = "text/javascript"
	MIMEApplicationXML  = "application/xml"
	MIMEApplicationJSON = "application/json"
	// Deprecated: use MIMETextJavaScript instead
	MIMEApplicationJavaScript = "application/javascript"
	MIMEApplicationForm       = "application/x-www-form-urlencoded"
	MIMEOctetStream           = "application/octet-stream"
	MIMEMultipartForm         = "multipart/form-data"

	MIMETextXMLCharsetUTF8         = "text/xml; charset=utf-8"
	MIMETextHTMLCharsetUTF8        = "text/html; charset=utf-8"
	MIMETextPlainCharsetUTF8       = "text/plain; charset=utf-8"
	MIMETextJavaScriptCharsetUTF8  = "text/javascript; charset=utf-8"
	MIMEApplicationXMLCharsetUTF8  = "application/xml; charset=utf-8"
	MIMEApplicationJSONCharsetUTF8 = "application/json; charset=utf-8"
	// Deprecated: use MIMETextJavaScriptCharsetUTF8 instead
	MIMEApplicationJavaScriptCharsetUTF8 = "application/javascript; charset=utf-8"
)

// header
const (
	HeaderAuthorization = "Authorization"
	HeaderHost          = "Host"
	HeaderReferer       = "Referer"
	HeaderContentType   = "Content-Type"
	HeaderUserAgent     = "User-Agent"
	HeaderExpect        = "Expect"
	HeaderConnection    = "Connection"
	HeaderContentLength = "Content-Length"
	HeaderCookie        = "Cookie"

	HeaderServer           = "Server"
	HeaderServerLower      = "server"
	HeaderSetCookie        = "Set-Cookie"
	HeaderSetCookieLower   = "set-cookie"
	HeaderTransferEncoding = "Transfer-Encoding"
	HeaderDate             = "Date"

	HeaderRange        = "Range"
	HeaderAcceptRanges = "Accept-Ranges"
	HeaderContentRange = "Content-Range"

	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderLastModified    = "Last-Modified"

	HeaderCacheControl       = "Cache-Control"
	HeaderContentDisposition = "Content-Disposition"

	// Message body information
	HeaderContentEncoding = "Content-Encoding"
	HeaderAcceptEncoding  = "Accept-Encoding"

	// Redirects
	HeaderLocation = "Location"

	// Protocol
	HTTP11 = "HTTP/1.1"
	HTTP10 = "HTTP/1.0"
	HTTP20 = "HTTP/2.0"
)
