// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package odatasql

import (
	"context"
	"errors"
	"fmt"

	"github.com/CiscoM31/godata"
)

type selectNode struct {
	children       map[string]*selectNode
	expandChildren map[string]struct{}
	selectChildren map[string]struct{}
	filter         *godata.GoDataFilterQuery
	orderby        *godata.GoDataOrderByQuery
	expand         bool
}

func newSelectTree() *selectNode {
	return &selectNode{
		children:       map[string]*selectNode{},
		expandChildren: map[string]struct{}{},
		selectChildren: map[string]struct{}{},
	}
}

func (st *selectNode) clone() *selectNode {
	return &selectNode{
		children:       st.children,
		expandChildren: st.expandChildren,
		selectChildren: st.selectChildren,
		filter:         st.filter,
		orderby:        st.orderby,
		expand:         st.expand,
	}
}

// nolint:gocognit,cyclop
func (st *selectNode) insert(path []*godata.Token, filter *godata.GoDataFilterQuery, orderby *godata.GoDataOrderByQuery, sel *godata.GoDataSelectQuery, subExpand *godata.GoDataExpandQuery, expand bool) error {
	// If path length == 0 then we've reach the bottom of the path, now we
	// need to save the filter/select and process any sub selects/expands
	if len(path) == 0 {
		if st.filter != nil {
			return errors.New("filter for field specified twice")
		}
		st.filter = filter

		if st.orderby != nil {
			return errors.New("orderby for field specified twice")
		}
		st.orderby = orderby

		st.expand = expand

		if sel != nil {
			if len(st.children) > 0 {
				return errors.New("can not specify selection for field in multiple places")
			}

			// Parse $select using ParseExpandString because godata.ParseSelectString
			// is a naive implementation and doesn't handle query options properly
			childSelections, err := godata.ParseExpandString(context.TODO(), sel.RawValue)
			if err != nil {
				return fmt.Errorf("failed to parse select: %w", err)
			}

			for _, s := range childSelections.ExpandItems {
				if s.Expand != nil {
					return errors.New("expand can not be specified inside of select")
				}
				err := st.insert(s.Path, s.Filter, s.OrderBy, s.Select, nil, false)
				if err != nil {
					return err
				}
			}
		}

		if subExpand != nil {
			for _, s := range subExpand.ExpandItems {
				err := st.insert(s.Path, s.Filter, s.OrderBy, s.Select, s.Expand, true)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	// Keep adding the items in the path to the tree. First find if this
	// entry is already in the tree by checking children, if it is then
	// insert the next part of the path into that child, otherwise add a
	// new child and insert it there.
	childName := path[0].Value
	child, ok := st.children[childName]
	if !ok {
		st.children[childName] = newSelectTree()
		child = st.children[childName]
	}

	// Keep track of which children where added as part of $select or
	// $expand, or both.
	if expand {
		st.expandChildren[childName] = struct{}{}
	} else {
		st.selectChildren[childName] = struct{}{}
	}

	err := child.insert(path[1:], filter, orderby, sel, subExpand, expand)
	if err != nil {
		return err
	}

	return nil
}
