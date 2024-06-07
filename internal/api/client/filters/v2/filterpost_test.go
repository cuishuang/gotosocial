// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package v2_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	filtersV2 "github.com/superseriousbusiness/gotosocial/internal/api/client/filters/v2"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func (suite *FiltersTestSuite) postFilter(title *string, context *[]string, action *string, expiresIn *int, requestJson *string, expectedHTTPStatus int, expectedBody string) (*apimodel.FilterV2, error) {
	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// create the request
	ctx.Request = httptest.NewRequest(http.MethodPost, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.BasePath, nil)
	ctx.Request.Header.Set("accept", "application/json")
	if requestJson != nil {
		ctx.Request.Header.Set("content-type", "application/json")
		ctx.Request.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		ctx.Request.Form = make(url.Values)
		if title != nil {
			ctx.Request.Form["title"] = []string{*title}
		}
		if context != nil {
			ctx.Request.Form["context[]"] = *context
		}
		if action != nil {
			ctx.Request.Form["filter_action"] = []string{*action}
		}
		if expiresIn != nil {
			ctx.Request.Form["expires_in"] = []string{strconv.Itoa(*expiresIn)}
		}
	}

	// trigger the handler
	suite.filtersModule.FilterPOSTHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
		if expectedBody == "" {
			return nil, errs.Combine()
		}
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.FilterV2{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestPostFilterFull() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	title := "GNU/Linux"
	context := []string{"home", "public"}
	action := "warn"
	expiresIn := 86400
	filter, err := suite.postFilter(&title, &context, &action, &expiresIn, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(title, filter.Title)
	filterContext := make([]string, 0, len(filter.Context))
	for _, c := range filter.Context {
		filterContext = append(filterContext, string(c))
	}
	suite.ElementsMatch(context, filterContext)
	suite.Equal(apimodel.FilterActionWarn, filter.FilterAction)
	if suite.NotNil(filter.ExpiresAt) {
		suite.NotEmpty(*filter.ExpiresAt)
	}
	suite.Empty(filter.Keywords)
	suite.Empty(filter.Statuses)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterFullJSON() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	// Use a numeric literal with a fractional part to test the JSON-specific handling for non-integer "expires_in".
	requestJson := `{
		"title": "GNU/Linux",
		"context": ["home", "public"],
		"filter_action": "warn",
		"whole_word": true,
		"expires_in": 86400.1
	}`
	filter, err := suite.postFilter(nil, nil, nil, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("GNU/Linux", filter.Title)
	suite.ElementsMatch(
		[]apimodel.FilterContext{
			apimodel.FilterContextHome,
			apimodel.FilterContextPublic,
		},
		filter.Context,
	)
	suite.Equal(apimodel.FilterActionWarn, filter.FilterAction)
	if suite.NotNil(filter.ExpiresAt) {
		suite.NotEmpty(*filter.ExpiresAt)
	}
	suite.Empty(filter.Keywords)
	suite.Empty(filter.Statuses)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterMinimal() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	title := "GNU/Linux"
	context := []string{"home"}
	filter, err := suite.postFilter(&title, &context, nil, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(title, filter.Title)
	filterContext := make([]string, 0, len(filter.Context))
	for _, c := range filter.Context {
		filterContext = append(filterContext, string(c))
	}
	suite.ElementsMatch(context, filterContext)
	suite.Equal(apimodel.FilterActionWarn, filter.FilterAction)
	suite.Nil(filter.ExpiresAt)
	suite.Empty(filter.Keywords)
	suite.Empty(filter.Statuses)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPostFilterEmptyTitle() {
	title := ""
	context := []string{"home"}
	_, err := suite.postFilter(&title, &context, nil, nil, nil, http.StatusUnprocessableEntity, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterMissingTitle() {
	context := []string{"home"}
	_, err := suite.postFilter(nil, &context, nil, nil, nil, http.StatusUnprocessableEntity, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterEmptyContext() {
	title := "GNU/Linux"
	context := []string{}
	_, err := suite.postFilter(&title, &context, nil, nil, nil, http.StatusUnprocessableEntity, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPostFilterMissingContext() {
	title := "GNU/Linux"
	_, err := suite.postFilter(&title, nil, nil, nil, nil, http.StatusUnprocessableEntity, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Creating another filter with the same title should fail.
func (suite *FiltersTestSuite) TestPostFilterTitleConflict() {
	title := suite.testFilters["local_account_1_filter_1"].Title
	_, err := suite.postFilter(&title, nil, nil, nil, nil, http.StatusUnprocessableEntity, "")
	if err != nil {
		suite.FailNow(err.Error())
	}
}
