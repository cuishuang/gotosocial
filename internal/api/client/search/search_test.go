/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package search_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type SearchStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	mediaManager media.Manager
	federator    federation.Federator
	processor    *processing.Processor
	emailSender  email.Sender
	sentEmails   map[string]string

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	// module being tested
	searchModule *search.Module
}

func (suite *SearchStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *SearchStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)

	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewInMemoryStorage()
	suite.mediaManager = testrig.NewTestMediaManager(suite.db, suite.storage)
	suite.federator = testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker), suite.storage, suite.mediaManager, fedWorker)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", suite.sentEmails)
	suite.processor = testrig.NewTestProcessor(suite.db, suite.storage, suite.federator, suite.emailSender, suite.mediaManager, clientWorker, fedWorker)
	suite.searchModule = search.New(suite.processor)
	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.NoError(suite.processor.Start())
}

func (suite *SearchStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func (suite *SearchStandardTestSuite) newContext(recorder *httptest.ResponseRecorder, requestPath string) *gin.Context {
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	protocol := config.GetProtocol()
	host := config.GetHost()

	baseURI := fmt.Sprintf("%s://%s", protocol, host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, requestPath)

	ctx.Request = httptest.NewRequest(http.MethodGet, requestURI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "application/json")
	return ctx
}
