package builder

import (
	"fmt"
	"strings"
	"time"
)

type GithubQueryParam interface {
	fmt.Stringer
}

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

type excluded struct {
	excluded bool
}

func (e *excluded) exclude() {
	e.excluded = true
}
func (e excluded) String() string {
	if e.excluded {
		return "-"
	} else {
		return ""
	}
}

type tagQuery struct {
	tag string
}

type textParam struct {
	excluded
	value string
}

func (t *textParam) Excluded() *textParam {
	t.exclude()
	return t
}
func (t *textParam) String() string {
	return fmt.Sprintf(`%s"%s"`, t.excluded, t.value)
}

func Text(value string) *textParam {
	return &textParam{
		value: value,
	}
}

type singleQueryParam struct {
	tagQuery
	excluded
	value string
}

func (s *singleQueryParam) String() string {
	return fmt.Sprintf("%s%s:%s", s.excluded, s.tag, s.value)
}

func (c *singleQueryParam) Excluded() *singleQueryParam {
	c.exclude()
	return c
}

func NewSingleQueryParam(tag string, value string) *singleQueryParam {
	return &singleQueryParam{
		tagQuery: tagQuery{tag: tag},
		value:    value,
	}
}

func Organization(orgName string) *singleQueryParam {
	return NewSingleQueryParam(tag_organization, orgName)
}

func Label(labels ...string) *singleQueryParam {
	return NewSingleQueryParam(tag_label, strings.Join(labels, ","))
}

func Repository(repoName string) *singleQueryParam {
	return NewSingleQueryParam(tag_repository, repoName)
}

type timeQueryBetween struct {
	tagQuery
	excluded
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
	c.exclude()
	return c
}

type timeQuerySingle struct {
	tagQuery
	excluded
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

func (c *timeQuerySingle) exclude() *timeQuerySingle {
	c.exclude()
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
