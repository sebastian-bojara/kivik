package memory

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/util"
	"github.com/go-kivik/mango"
)

var errFindNotImplemented = errors.New("find feature not yet implemented")

type findQuery struct {
	Selector *mango.Selector `json:"selector"`
	Limit    int64           `json:"limit"`
	Skip     int64           `json:"skip"`
	Sort     []string        `json:"sort"`
	Fields   []string        `json:"fields"`
	UseIndex indexSpec       `json:"use_index"`
}

type indexSpec struct {
	ddoc  string
	index string
}

func (i *indexSpec) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &i.ddoc)
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if len(values) == 0 || len(values) > 2 {
		return errors.New("invalid index specification")
	}
	i.ddoc = values[0]
	if len(values) == 2 {
		i.index = values[1]
	}
	return nil
}

func (d *db) CreateIndex(_ context.Context, ddoc, name string, index interface{}) error {
	return errFindNotImplemented
}

func (d *db) GetIndexes(_ context.Context) ([]driver.Index, error) {
	return nil, errFindNotImplemented
}

func (d *db) DeleteIndex(_ context.Context, ddoc, name string) error {
	return errFindNotImplemented
}

func (d *db) Find(_ context.Context, query interface{}) (driver.Rows, error) {
	queryJSON, err := util.ToJSON(query)
	if err != nil {
		return nil, err
	}
	fq := &findQuery{}
	if err := json.NewDecoder(queryJSON).Decode(&fq); err != nil {
		return nil, err
	}
	if fq == nil || fq.Selector == nil {
		return nil, errors.New("Missing required key: selector")
	}
	rows := &findResults{
		resultSet{
			docIDs: make([]string, 0),
			revs:   make([]*revision, 0),
		},
	}
	for docID := range d.db.docs {
		if doc, found := d.db.latestRevision(docID); found {
			cd, err := toCouchDoc(doc)
			if err != nil {
				panic(err)
			}
			match, err := fq.Selector.Matches(map[string]interface{}(cd))
			if err != nil {
				return nil, err
			}
			if match {
				rows.docIDs = append(rows.docIDs, docID)
				rows.revs = append(rows.revs, doc)
			}
		}
	}
	rows.offset = 0
	rows.totalRows = int64(len(rows.docIDs))
	return rows, nil
}

type findResults struct {
	resultSet
}

var _ driver.Rows = &findResults{}

func (r *findResults) Next(row *driver.Row) error {
	if r.revs == nil || len(r.revs) == 0 {
		return io.EOF
	}
	row.ID, r.docIDs = r.docIDs[0], r.docIDs[1:]
	row.Doc = r.revs[0].data
	r.revs = r.revs[1:]
	return nil
}
