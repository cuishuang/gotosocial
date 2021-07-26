/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package text

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Formatter wraps some logic and functions for parsing statuses and other text input into nice html.
type Formatter interface {
	// FromMarkdown parses an HTML text from a markdown-formatted text.
	FromMarkdown(md string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string
	// FromPlain parses an HTML text from a plaintext.
	FromPlain(plain string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string
}

type formatter struct {
	cfg *config.Config
	db  db.DB
	log *logrus.Logger
}

// NewFormatter returns a new Formatter interface for parsing statuses and other text input into nice html.
func NewFormatter(cfg *config.Config, db db.DB, log *logrus.Logger) Formatter {
	return &formatter{
		cfg: cfg,
		db:  db,
		log: log,
	}
}
