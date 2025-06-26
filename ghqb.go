package ghqb

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type timeOrd int8

const (
	ORD_EQ  timeOrd = 1 << iota
	ORD_LT  timeOrd = 1 << iota
	ORD_GT  timeOrd = 1 << iota
	ORD_GEQ timeOrd = ORD_EQ | ORD_LT
	ORD_LEQ timeOrd = ORD_EQ | ORD_GT
)

type QueryType string

const (
	QueryText        QueryType = "text"
	QueryTag         QueryType = "tag"
	QueryTimeBetween QueryType = "time_between"
	QueryTime        QueryType = "time"
)

const (
	TagLabel        string = "label"
	TagAuthor       string = "author"
	TagRepository   string = "repo"
	TagOrganization string = "org"
	TagCreated      string = "created"
	TagClosed       string = "closed"
)

const (
	Separator string = ","
)

type GithubQuery interface {
	Format() (string, error)
	QueryType() QueryType
}

func (t timeOrd) String() string {
	var out = ""
	switch t {
	case ORD_EQ:
		out = ""
	case ORD_LT:
		out = "<"
	case ORD_LEQ:
		out = "<="
	case ORD_GEQ:
		out = ">="
	case ORD_GT:
		out = ">"
	}
	return out
}

type baseQuery struct {
	Excluded  bool
	queryType QueryType
}

func (e *baseQuery) exclude() {
	e.Excluded = true
}
func (t *baseQuery) QueryType() QueryType {
	return t.queryType
}
func (e baseQuery) String() string {
	if e.Excluded {
		return "-"
	} else {
		return ""
	}
}

type TextQuery struct {
	baseQuery
	Value string
}

func (t *TextQuery) Exclude() *TextQuery {
	t.exclude()
	return t
}
func (t *TextQuery) Format() (string, error) {
	return fmt.Sprintf(`%s"%s"`, t.baseQuery, t.Value), nil
}

func Text(value string) *TextQuery {
	return &TextQuery{
		Value:     value,
		baseQuery: baseQuery{queryType: QueryText},
	}
}

type TagQuery struct {
	baseQuery
	Tag   string
	Value string
}

func (s *TagQuery) Format() (string, error) {
	return fmt.Sprintf("%s%s:%s", s.baseQuery, s.Tag, s.Value), nil
}

func (c *TagQuery) Exclude() *TagQuery {
	c.exclude()
	return c
}

func Tag(tag string, value string) *TagQuery {
	return &TagQuery{
		Tag:       tag,
		Value:     value,
		baseQuery: baseQuery{queryType: QueryTag},
	}
}

func Organization(orgName string) *TagQuery {
	return Tag(TagOrganization, orgName)
}

func Label(labels ...string) *TagQuery {
	return Tag(TagLabel, strings.Join(labels, Separator))
}

func Author(authors ...string) *TagQuery {
	return Tag(TagAuthor, strings.Join(authors, Separator))
}

func Repository(repoName string) *TagQuery {
	return Tag(TagRepository, repoName)
}

type TimeQueryBetween struct {
	baseQuery
	Tag        string
	TimeFormat string
	Start      time.Time
	End        time.Time
}

func (c *TimeQueryBetween) Format() (string, error) {
	if c.Start.After(c.End) {
		return "", ErrInvalidTimePeriod
	}
	return fmt.Sprintf(
		"%s%s:%s..%s",
		c.baseQuery,
		c.Tag,
		c.Start.Format(c.TimeFormat),
		c.End.Format(c.TimeFormat),
	), nil
}

func (c *TimeQueryBetween) Exclude() *TimeQueryBetween {
	c.exclude()
	return c
}

type TimeQuery struct {
	baseQuery
	Tag        string
	TimeFormat string
	Ord        timeOrd
	Value      time.Time
}

func (t *TimeQuery) Format() (string, error) {
	return fmt.Sprintf(
		"%s%s:%s%s",
		t.baseQuery,
		t.Tag,
		t.Ord,
		t.Value.Format(t.TimeFormat),
	), nil
}

func (c *TimeQuery) Exclude() *TimeQuery {
	c.exclude()
	return c
}

func Between(tag, format string, start, end time.Time) *TimeQueryBetween {
	return &TimeQueryBetween{
		Tag:        tag,
		TimeFormat: format,
		Start:      start,
		End:        end,
		baseQuery:  baseQuery{queryType: QueryTimeBetween},
	}

}

func CreatedBetween(start, end time.Time) *TimeQueryBetween {
	return Between(TagCreated, time.DateOnly, start, end)
}
func ClosedBetween(start, end time.Time) *TimeQueryBetween {
	return Between(TagClosed, time.DateOnly, start, end)
}
func CreatedBetweenTimezoned(start, end time.Time) *TimeQueryBetween {
	return Between(TagCreated, time.RFC3339, start, end)
}
func ClosedBetweenTimezoned(start, end time.Time) *TimeQueryBetween {
	return Between(TagClosed, time.RFC3339, start, end)
}

func Time(tag, format string, t time.Time, ord timeOrd) *TimeQuery {
	return &TimeQuery{
		Tag:        tag,
		TimeFormat: format,
		Value:      t,
		Ord:        ord,
		baseQuery:  baseQuery{queryType: QueryTime},
	}
}

func Created(t time.Time, ord timeOrd) *TimeQuery {
	return Time(TagCreated, time.DateOnly, t, ord)
}
func Closed(t time.Time, ord timeOrd) *TimeQuery {
	return Time(TagClosed, time.DateOnly, t, ord)
}
func CreatedTimezoned(t time.Time, ord timeOrd) *TimeQuery {
	return Time(TagCreated, time.RFC3339, t, ord)
}
func ClosedTimezoned(t time.Time, ord timeOrd) *TimeQuery {
	return Time(TagClosed, time.RFC3339, t, ord)
}

func Query(params ...GithubQuery) (string, error) {
	var builder strings.Builder
	var errs []error
	for _, param := range params {
		queryPiece, err := param.Format()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		builder.WriteString(queryPiece)
		builder.WriteRune(' ')
	}
	return builder.String(), errors.Join(errs...)
}
