package builder

import (
	"fmt"
	"strings"
	"time"
)

type GithubQueryParam interface {
	fmt.Stringer
}

const (
	GhQueryDateFormat = "2006-01-02T15:04:05+07:00"
)

type timeOrd int8

const (
	ORD_EQ  timeOrd = 1 << iota
	ORD_LT  timeOrd = 1 << iota
	ORD_GT  timeOrd = 1 << iota
	ORD_GEQ timeOrd = ORD_EQ | ORD_LT
	ORD_LEQ timeOrd = ORD_EQ | ORD_GT
)

const (
	tag_label        string = "label"
	tag_repository   string = "repo"
	tag_organization string = "org"
)

func (t timeOrd) String() string {
	var out = ""
	switch t {
	case ORD_LT:
		out = "<"
	case ORD_LEQ:
		out = "<="
	case ORD_EQ:
		out = "="
	case ORD_GEQ:
		out = ">="
	case ORD_GT:
		out = ">"
	}
	return out
}

type queryString string

func (s queryString) String() string {
	return string(s)
}

type excludedQuery bool

type tagQuery struct {
	excluded excludedQuery
	tag      string
}

func (s *tagQuery) Exclude() {
	s.excluded = true
}

func (t excludedQuery) String() string {
	if t {
		return "-"
	} else {
		return ""
	}
}

type singleQueryParam[T fmt.Stringer] struct {
	tagQuery
	value T
}

func (s *singleQueryParam[T]) String() string {
	return fmt.Sprintf("%s%s:%s", s.excluded, s.tag, s.value)
}

func (c *singleQueryParam[T]) Excluded() *singleQueryParam[T] {
	c.Exclude()
	return c
}

func NewSingleQueryParam[T fmt.Stringer](tag string, value T) *singleQueryParam[T] {
	return &singleQueryParam[T]{
		tagQuery: tagQuery{tag: tag},
		value:    value,
	}
}

func Organization(orgName string) *singleQueryParam[queryString] {
	return NewSingleQueryParam(tag_organization, queryString(orgName))
}

func Label(labels ...string) *singleQueryParam[queryString] {
	return NewSingleQueryParam(tag_label, queryString(strings.Join(labels, ",")))
}

func Repository(repoName string) *singleQueryParam[queryString] {
	return NewSingleQueryParam(tag_repository, queryString(repoName))
}

type timeQueryBetween struct {
	tagQuery
	format string
	start  time.Time
	end    time.Time
}

func (c *timeQueryBetween) String() string {
	return fmt.Sprintf(
		"%s%s:%s..%s",
		c.excluded,
		c.tag,
		c.start.Format(c.format),
		c.end.Format(c.format),
	)
}

func (c *timeQueryBetween) Excluded() *timeQueryBetween {
	c.Exclude()
	return c
}

type timeQuerySingle struct {
	tagQuery
	format string
	ord    timeOrd
	value  time.Time
}

func (t *timeQuerySingle) String() string {
	return fmt.Sprintf(
		"%s%s:%s%s",
		t.excluded,
		t.tag,
		t.ord,
		t.value.Format(t.format),
	)
}

func (c *timeQuerySingle) Exclude() *timeQuerySingle {
	c.Exclude()
	return c
}

func buildTimeBetween(tag, format string, start, end time.Time) *timeQueryBetween {
	return &timeQueryBetween{
		tagQuery: tagQuery{tag: tag},
		format:   format,
		start:    start,
		end:      end,
	}

}

func CreatedBetween(start, end time.Time) *timeQueryBetween {
	return buildTimeBetween("created", time.DateOnly, start, end)
}
func ClosedBetween(start, end time.Time) *timeQueryBetween {
	return buildTimeBetween("closed", time.DateOnly, start, end)
}
func CreatedBetweenTimezoned(start, end time.Time) *timeQueryBetween {
	return buildTimeBetween("created", time.RFC3339, start, end)
}
func ClosedBetweenTimezoned(start, end time.Time) *timeQueryBetween {
	return buildTimeBetween("closed", time.RFC3339, start, end)
}

func buildTimeOrd(tag, format string, t time.Time, ord timeOrd) *timeQuerySingle {
	return &timeQuerySingle{
		tagQuery: tagQuery{tag: tag},
		format:   format,
		value:    t,
		ord:      ord,
	}
}

func Created(t time.Time, ord timeOrd) *timeQuerySingle {
	return buildTimeOrd("created", time.DateOnly, t, ord)
}
func Closed(t time.Time, ord timeOrd) *timeQuerySingle {
	return buildTimeOrd("closed", time.DateOnly, t, ord)
}
func CreatedTimezoned(t time.Time, ord timeOrd) *timeQuerySingle {
	return buildTimeOrd("created", time.RFC3339, t, ord)
}
func ClosedTimezoned(t time.Time, ord timeOrd) *timeQuerySingle {
	return buildTimeOrd("closed", time.RFC3339, t, ord)
}

func Query(params ...GithubQueryParam) (string, error) {
	var builder strings.Builder
	var err error
	for _, param := range params {
		builder.WriteString(param.String())
		builder.WriteRune(' ')
	}
	return builder.String(), err
}
