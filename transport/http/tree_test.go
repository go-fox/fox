package http

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestAddRoute(t *testing.T) {
	hStub := Handler(func(ctx *Context) error { return nil })
	hIndex := Handler(func(ctx *Context) error { return nil })
	hFavicon := Handler(func(ctx *Context) error { return nil })
	hArticleList := Handler(func(ctx *Context) error { return nil })
	hArticleNear := Handler(func(ctx *Context) error { return nil })
	hArticleShow := Handler(func(ctx *Context) error { return nil })
	hArticleShowRelated := Handler(func(ctx *Context) error { return nil })
	hArticleShowOpts := Handler(func(ctx *Context) error { return nil })
	hArticleSlug := Handler(func(ctx *Context) error { return nil })
	hArticleByUser := Handler(func(ctx *Context) error { return nil })
	hUserList := Handler(func(ctx *Context) error { return nil })
	hUserShow := Handler(func(ctx *Context) error { return nil })
	hAdminCatchall := Handler(func(ctx *Context) error { return nil })
	hAdminAppShow := Handler(func(ctx *Context) error { return nil })
	hAdminAppShowCatchall := Handler(func(ctx *Context) error { return nil })
	hUserProfile := Handler(func(ctx *Context) error { return nil })
	hUserSuper := Handler(func(ctx *Context) error { return nil })
	hUserAll := Handler(func(ctx *Context) error { return nil })
	hHubView1 := Handler(func(ctx *Context) error { return nil })
	hHubView2 := Handler(func(ctx *Context) error { return nil })
	hHubView3 := Handler(func(ctx *Context) error { return nil })

	tr := &node{}

	tr.AddRoute(mGET, "/", hIndex)
	tr.AddRoute(mGET, "/favicon.ico", hFavicon)

	tr.AddRoute(mGET, "/pages/*", hStub)

	tr.AddRoute(mGET, "/article", hArticleList)
	tr.AddRoute(mGET, "/article/", hArticleList)

	tr.AddRoute(mGET, "/article/near", hArticleNear)
	tr.AddRoute(mGET, "/article/{id}", hStub)
	tr.AddRoute(mGET, "/article/{id}", hArticleShow)
	tr.AddRoute(mGET, "/article/{id}", hArticleShow) // duplicate will have no effect
	tr.AddRoute(mGET, "/article/@{user}", hArticleByUser)

	tr.AddRoute(mGET, "/article/{sup}/{opts}", hArticleShowOpts)
	tr.AddRoute(mGET, "/article/{id}/{opts}", hArticleShowOpts) // overwrite above route, latest wins

	tr.AddRoute(mGET, "/article/{iffd}/edit", hStub)
	tr.AddRoute(mGET, "/article/{id}//related", hArticleShowRelated)
	tr.AddRoute(mGET, "/article/slug/{month}/-/{day}/{year}", hArticleSlug)

	tr.AddRoute(mGET, "/admin/user", hUserList)
	tr.AddRoute(mGET, "/admin/user/", hStub) // will get replaced by next route
	tr.AddRoute(mGET, "/admin/user/", hUserList)

	tr.AddRoute(mGET, "/admin/user//{id}", hUserShow)
	tr.AddRoute(mGET, "/admin/user/{id}", hUserShow)

	tr.AddRoute(mGET, "/admin/apps/{id}", hAdminAppShow)
	tr.AddRoute(mGET, "/admin/apps/{id}/*", hAdminAppShowCatchall)

	tr.AddRoute(mGET, "/admin/*", hStub) // catchall segment will get replaced by next route
	tr.AddRoute(mGET, "/admin/*", hAdminCatchall)

	tr.AddRoute(mGET, "/users/{userID}/profile", hUserProfile)
	tr.AddRoute(mGET, "/users/super/*", hUserSuper)
	tr.AddRoute(mGET, "/users/*", hUserAll)

	tr.AddRoute(mGET, "/hubs/{hubID}/view", hHubView1)
	tr.AddRoute(mGET, "/hubs/{hubID}/view/*", hHubView2)
	//sr := NewRouter()
	//sr.Get("/users", hHubView3)
	//tr.AddRoute(mGET, "/hubs/{hubID}/*", sr)
	tr.AddRoute(mGET, "/hubs/{hubID}/users", hHubView3)
	srv := Server{}
	ctx := srv.acquireContext(&fasthttp.RequestCtx{})
	route, _, _ := tr.FindRoute(ctx, mGET, "/favicon.ico")
	if route == nil {
		t.Error("route not found")
		return
	}
	handler := route.endpoints.Value(mGET).handler
	if handler == nil {
		t.Error("handler not found")
		return
	}
	if fmt.Sprintf("%v", hFavicon) != fmt.Sprintf("%v", handler) {
		t.Error("handler mismatch")
		return
	}
	println(ctx.routeParams.keys, ctx.routeParams.values)
}

func TestTree(t *testing.T) {
	hStub := Handler(func(ctx *Context) error { return nil })
	//hIndex := Handler(func(ctx *Context) error { return nil })
	//hFavicon := Handler(func(ctx *Context) error { return nil })
	//hArticleList := Handler(func(ctx *Context) error { return nil })
	//hArticleNear := Handler(func(ctx *Context) error { return nil })
	hArticleShow := Handler(func(ctx *Context) error { return nil })
	//hArticleShowRelated := Handler(func(ctx *Context) error { return nil })
	//hArticleShowOpts := Handler(func(ctx *Context) error { return nil })
	//hArticleSlug := Handler(func(ctx *Context) error { return nil })
	//hArticleByUser := Handler(func(ctx *Context) error { return nil })
	//hUserList := Handler(func(ctx *Context) error { return nil })
	//hUserShow := Handler(func(ctx *Context) error { return nil })
	//hAdminCatchall := Handler(func(ctx *Context) error { return nil })
	//hAdminAppShow := Handler(func(ctx *Context) error { return nil })
	//hAdminAppShowCatchall := Handler(func(ctx *Context) error { return nil })
	//hUserProfile := Handler(func(ctx *Context) error { return nil })
	//hUserSuper := Handler(func(ctx *Context) error { return nil })
	//hUserAll := Handler(func(ctx *Context) error { return nil })
	//hHubView1 := Handler(func(ctx *Context) error { return nil })
	//hHubView2 := Handler(func(ctx *Context) error { return nil })
	//hHubView3 := Handler(func(ctx *Context) error { return nil })

	tr := &node{}

	//tr.AddRoute(mGET, "/", hIndex)
	//tr.AddRoute(mGET, "/favicon.ico", hFavicon)
	//
	//tr.AddRoute(mGET, "/pages/*", hStub)
	//
	//tr.AddRoute(mGET, "/article", hArticleList)
	//tr.AddRoute(mGET, "/article/", hArticleList)
	//
	//tr.AddRoute(mGET, "/article/near", hArticleNear)
	tr.AddRoute(mGET, "/article/{id}", hStub)
	tr.AddRoute(mGET, "/article/{id}", hArticleShow)
	//tr.AddRoute(mGET, "/article/{id}", hArticleShow) // duplicate will have no effect
	//tr.AddRoute(mGET, "/article/@{user}", hArticleByUser)
	//
	//tr.AddRoute(mGET, "/article/{sup}/{opts}", hArticleShowOpts)
	//tr.AddRoute(mGET, "/article/{id}/{opts}", hArticleShowOpts) // overwrite above route, latest wins
	//
	//tr.AddRoute(mGET, "/article/{iffd}/edit", hStub)
	//tr.AddRoute(mGET, "/article/{id}//related", hArticleShowRelated)
	//tr.AddRoute(mGET, "/article/slug/{month}/-/{day}/{year}", hArticleSlug)
	//
	//tr.AddRoute(mGET, "/admin/user", hUserList)
	//tr.AddRoute(mGET, "/admin/user/", hStub) // will get replaced by next route
	//tr.AddRoute(mGET, "/admin/user/", hUserList)
	//
	//tr.AddRoute(mGET, "/admin/user//{id}", hUserShow)
	//tr.AddRoute(mGET, "/admin/user/{id}", hUserShow)
	//
	//tr.AddRoute(mGET, "/admin/apps/{id}", hAdminAppShow)
	//tr.AddRoute(mGET, "/admin/apps/{id}/*", hAdminAppShowCatchall)
	//
	//tr.AddRoute(mGET, "/admin/*", hStub) // catchall segment will get replaced by next route
	//tr.AddRoute(mGET, "/admin/*", hAdminCatchall)
	//
	//tr.AddRoute(mGET, "/users/{userID}/profile", hUserProfile)
	//tr.AddRoute(mGET, "/users/super/*", hUserSuper)
	//tr.AddRoute(mGET, "/users/*", hUserAll)
	//
	//tr.AddRoute(mGET, "/hubs/{hubID}/view", hHubView1)
	//tr.AddRoute(mGET, "/hubs/{hubID}/view/*", hHubView2)
	////sr := NewRouter()
	////sr.Get("/users", hHubView3)
	////tr.AddRoute(mGET, "/hubs/{hubID}/*", sr)
	//tr.AddRoute(mGET, "/hubs/{hubID}/users", hHubView3)

	tests := []struct {
		r string   // input request path
		h Handler  // output matched handler
		k []string // output param keys
		v []string // output param values
	}{
		//{r: "/", h: hIndex, k: []string{}, v: []string{}},
		//{r: "/favicon.ico", h: hFavicon, k: []string{}, v: []string{}},
		//
		//{r: "/pages", h: nil, k: []string{}, v: []string{}},
		//{r: "/pages/", h: hStub, k: []string{"*"}, v: []string{""}},
		//{r: "/pages/yes", h: hStub, k: []string{"*"}, v: []string{"yes"}},
		//
		//{r: "/article", h: hArticleList, k: []string{}, v: []string{}},
		//{r: "/article/", h: hArticleList, k: []string{}, v: []string{}},
		//{r: "/article/near", h: hArticleNear, k: []string{}, v: []string{}},
		//{r: "/article/neard", h: hArticleShow, k: []string{"id"}, v: []string{"neard"}},
		{r: "/article/123", h: hArticleShow, k: []string{"id"}, v: []string{"123"}},
		//{r: "/article/123/456", h: hArticleShowOpts, k: []string{"id", "opts"}, v: []string{"123", "456"}},
		//{r: "/article/@peter", h: hArticleByUser, k: []string{"user"}, v: []string{"peter"}},
		//{r: "/article/22//related", h: hArticleShowRelated, k: []string{"id"}, v: []string{"22"}},
		//{r: "/article/111/edit", h: hStub, k: []string{"iffd"}, v: []string{"111"}},
		//{r: "/article/slug/sept/-/4/2015", h: hArticleSlug, k: []string{"month", "day", "year"}, v: []string{"sept", "4", "2015"}},
		//{r: "/article/:id", h: hArticleShow, k: []string{"id"}, v: []string{":id"}},
		//
		//{r: "/admin/user", h: hUserList, k: []string{}, v: []string{}},
		//{r: "/admin/user/", h: hUserList, k: []string{}, v: []string{}},
		//{r: "/admin/user/1", h: hUserShow, k: []string{"id"}, v: []string{"1"}},
		//{r: "/admin/user//1", h: hUserShow, k: []string{"id"}, v: []string{"1"}},
		//{r: "/admin/hi", h: hAdminCatchall, k: []string{"*"}, v: []string{"hi"}},
		//{r: "/admin/lots/of/:fun", h: hAdminCatchall, k: []string{"*"}, v: []string{"lots/of/:fun"}},
		//{r: "/admin/apps/333", h: hAdminAppShow, k: []string{"id"}, v: []string{"333"}},
		//{r: "/admin/apps/333/woot", h: hAdminAppShowCatchall, k: []string{"id", "*"}, v: []string{"333", "woot"}},
		//
		//{r: "/hubs/123/view", h: hHubView1, k: []string{"hubID"}, v: []string{"123"}},
		//{r: "/hubs/123/view/index.html", h: hHubView2, k: []string{"hubID", "*"}, v: []string{"123", "index.html"}},
		//{r: "/hubs/123/users", h: hHubView3, k: []string{"hubID"}, v: []string{"123"}},
		//
		//{r: "/users/123/profile", h: hUserProfile, k: []string{"userID"}, v: []string{"123"}},
		//{r: "/users/super/123/okay/yes", h: hUserSuper, k: []string{"*"}, v: []string{"123/okay/yes"}},
		//{r: "/users/123/okay/yes", h: hUserAll, k: []string{"*"}, v: []string{"123/okay/yes"}},
	}

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, tr, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	for i, tt := range tests {
		srv := NewServer()
		ctx := srv.acquireContext(&fasthttp.RequestCtx{})

		route, _, _ := tr.FindRoute(ctx, mGET, "/favicon.ico")

		var handler Handler
		if methodHandler, ok := route.endpoints[mGET]; ok {
			handler = methodHandler.handler
		}

		paramKeys := ctx.routeParams.keys
		paramValues := ctx.routeParams.values

		if fmt.Sprintf("%v", tt.h) != fmt.Sprintf("%v", handler) {
			t.Errorf("input [%d]: find '%s' expecting handler:%v , got:%v", i, tt.r, tt.h, handler)
		}
		if !stringSliceEqual(tt.k, paramKeys) {
			t.Errorf("input [%d]: find '%s' expecting paramKeys:(%d)%v , got:(%d)%v", i, tt.r, len(tt.k), tt.k, len(paramKeys), paramKeys)
		}
		if !stringSliceEqual(tt.v, paramValues) {
			t.Errorf("input [%d]: find '%s' expecting paramValues:(%d)%v , got:(%d)%v", i, tt.r, len(tt.v), tt.v, len(paramValues), paramValues)
		}
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if b[i] != a[i] {
			return false
		}
	}
	return true
}
