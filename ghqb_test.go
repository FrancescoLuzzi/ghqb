package builder

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testTime = time.Now()
var timeOrds []timeOrd = []timeOrd{
	ORD_EQ,
	ORD_LT,
	ORD_LEQ,
	ORD_GEQ,
	ORD_GT,
}

const testOrgRepo = "test"

func addRepoCase(tests map[GithubQueryParam]string, repoName string) {
	tests[Repository(repoName)] = fmt.Sprintf("repo:%s", repoName)
}

func addOrgCase(tests map[GithubQueryParam]string, orgName string) {
	tests[Organization(orgName)] = fmt.Sprintf("org:%s", orgName)
}

func addCreatedCase(tests map[GithubQueryParam]string, t time.Time, o timeOrd) {
	tests[Created(t, o)] = fmt.Sprintf("created:%s%s", o, t.Format(time.DateOnly))
	tests[CreatedTimezoned(t, o)] = fmt.Sprintf("created:%s%s", o, t.Format(time.RFC3339))
}

func addClosedCase(tests map[GithubQueryParam]string, t time.Time, o timeOrd) {
	tests[Closed(t, o)] = fmt.Sprintf("closed:%s%s", o, t.Format(time.DateOnly))
	tests[ClosedTimezoned(t, o)] = fmt.Sprintf("closed:%s%s", o, t.Format(time.RFC3339))
}

func addCreatedBetween(tests map[GithubQueryParam]string, start, end time.Time) {
	tests[CreatedBetween(start, end)] = fmt.Sprintf("created:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[CreatedBetweenTimezoned(start, end)] = fmt.Sprintf("created:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
}
func addClosedBetween(tests map[GithubQueryParam]string, start, end time.Time) {
	tests[ClosedBetween(start, end)] = fmt.Sprintf("closed:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[ClosedBetweenTimezoned(start, end)] = fmt.Sprintf("closed:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

func TestOrdCases(t *testing.T) {
	tests := map[timeOrd]string{
		ORD_EQ:  "",
		ORD_LT:  "<",
		ORD_LEQ: "<=",
		ORD_GEQ: ">=",
		ORD_GT:  ">",
	}
	for test, expected := range tests {
		assert.Equal(t, expected, test.String())
	}
}

func TestQueryString(t *testing.T) {
	tests := make(map[GithubQueryParam]string)
	addRepoCase(tests, testOrgRepo)
	addOrgCase(tests, testOrgRepo)
	utcTestTime := testTime.UTC()
	for _, o := range timeOrds {
		addCreatedCase(tests, testTime, o)
		addClosedCase(tests, testTime, o)
		addCreatedCase(tests, utcTestTime, o)
		addClosedCase(tests, utcTestTime, o)
	}

	offset, err := rand.Int(rand.Reader, big.NewInt(120))
	assert.Nil(t, err, "something went wrong while generating a random offset")

	end := testTime.Add(time.Hour * time.Duration(offset.Int64()))
	addCreatedBetween(tests, testTime, end)
	addClosedBetween(tests, testTime, end)
	addCreatedBetween(tests, utcTestTime, end)
	addClosedBetween(tests, utcTestTime, end)

	for test, expected := range tests {
		assert.Equal(t, expected, test.String())
	}
}

var res GithubQueryParam

func Benchmark_Repository(b *testing.B) {
	var r GithubQueryParam
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		r = Repository(testOrgRepo)
	}
	res = r
}
