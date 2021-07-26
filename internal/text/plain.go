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
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *formatter) FromPlain(plain string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string {
	content := preformat(plain)

	// format mentions nicely
	for _, menchie := range mentions {
		targetAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(menchie.TargetAccountID, targetAccount); err == nil {
			mentionContent := fmt.Sprintf(`<span class="h-card"><a href="%s" class="u-url mention">@<span>%s</span></a></span>`, targetAccount.URL, targetAccount.Username)
			content = strings.ReplaceAll(content, menchie.NameString, mentionContent)
		}
	}

	// format tags nicely
	for _, tag := range tags {
		tagContent := fmt.Sprintf(`<a href="%s" class="mention hashtag" rel="tag">#<span>%s</span></a>`, tag.URL, tag.Name)
		content = strings.ReplaceAll(content, fmt.Sprintf("#%s", tag.Name), tagContent)
	}

	// replace newlines with breaks
	content = strings.ReplaceAll(content, "\n", "<br />")

	return postformat(content)
}
