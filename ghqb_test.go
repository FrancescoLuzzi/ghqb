package ghqb

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
	ORD_LT,
	ORD_LEQ,
	ORD_EQ,
	ORD_GEQ,
	ORD_GT,
}

const testOrgRepo = "test"

func addRepoCase(tests map[GithubQuery]string, repoName string) {
	tests[Repository(repoName)] = fmt.Sprintf("repo:%s", repoName)
	tests[Repository(repoName).Exclude()] = fmt.Sprintf("-repo:%s", repoName)
}

func addOrgCase(tests map[GithubQuery]string, orgName string) {
	tests[Organization(orgName)] = fmt.Sprintf("org:%s", orgName)
	tests[Organization(orgName).Exclude()] = fmt.Sprintf("-org:%s", orgName)
}

func addCreatedCase(tests map[GithubQuery]string, t time.Time, o timeOrd) {
	tests[Created(t, o)] = fmt.Sprintf("created:%s%s", o, t.Format(time.DateOnly))
	tests[Created(t, o).Exclude()] = fmt.Sprintf("-created:%s%s", o, t.Format(time.DateOnly))
	tests[CreatedTimezoned(t, o)] = fmt.Sprintf("created:%s%s", o, t.Format(time.RFC3339))
	tests[CreatedTimezoned(t, o).Exclude()] = fmt.Sprintf("-created:%s%s", o, t.Format(time.RFC3339))
}

func addClosedCase(tests map[GithubQuery]string, t time.Time, o timeOrd) {
	tests[Closed(t, o)] = fmt.Sprintf("closed:%s%s", o, t.Format(time.DateOnly))
	tests[Closed(t, o).Exclude()] = fmt.Sprintf("-closed:%s%s", o, t.Format(time.DateOnly))
	tests[ClosedTimezoned(t, o)] = fmt.Sprintf("closed:%s%s", o, t.Format(time.RFC3339))
	tests[ClosedTimezoned(t, o).Exclude()] = fmt.Sprintf("-closed:%s%s", o, t.Format(time.RFC3339))
}

func addCreatedBetweenCase(tests map[GithubQuery]string, start, end time.Time) {
	tests[CreatedBetween(start, end)] = fmt.Sprintf("created:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[CreatedBetween(start, end).Exclude()] = fmt.Sprintf("-created:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[CreatedBetweenTimezoned(start, end)] = fmt.Sprintf("created:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	tests[CreatedBetweenTimezoned(start, end).Exclude()] = fmt.Sprintf("-created:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

func addClosedBetweenCase(tests map[GithubQuery]string, start, end time.Time) {
	tests[ClosedBetween(start, end)] = fmt.Sprintf("closed:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[ClosedBetween(start, end).Exclude()] = fmt.Sprintf("-closed:%s..%s", start.Format(time.DateOnly), end.Format(time.DateOnly))
	tests[ClosedBetweenTimezoned(start, end)] = fmt.Sprintf("closed:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	tests[ClosedBetweenTimezoned(start, end).Exclude()] = fmt.Sprintf("-closed:%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
}

func addTextCase(tests map[GithubQuery]string, value string) {
	tests[Text(value)] = fmt.Sprintf(`"%s"`, value)
	tests[Text(value).Exclude()] = fmt.Sprintf(`-"%s"`, value)
}

func TestTimeOrd(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	tests := make(map[GithubQuery]string)
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
	addTextCase(tests, testOrgRepo)
	addCreatedBetweenCase(tests, testTime, end)
	addClosedBetweenCase(tests, testTime, end)
	addCreatedBetweenCase(tests, utcTestTime, end)
	addClosedBetweenCase(tests, utcTestTime, end)

	for test, expected := range tests {
		value, err := test.Format()
		assert.NoError(t, err)
		assert.Equal(t, expected, value)
	}
}

func TestQueryBetweenFail(t *testing.T) {
	now := time.Now()
	future := now.AddDate(1, 0, 0)
	type testFunc func(start, end time.Time) *TimeQueryBetween
	tests := []testFunc{
		CreatedBetween,
		ClosedBetween,
		ClosedBetweenTimezoned,
		CreatedBetweenTimezoned,
	}
	for _, f := range tests {
		_, err := f(future, now).Format()
		assert.ErrorIs(t, err, ErrInvalidTimePeriod)
	}
}

func TestQueryOK(t *testing.T) {
	now := time.Now()
	future := now.AddDate(1, 0, 0)
	_, err := Query(
		Text("test"),
		Repository("test"),
		Organization("test"),
		CreatedBetween(now, future),
	)
	assert.NoError(t, err)
}

func TestQueryKO(t *testing.T) {
	now := time.Now()
	future := now.AddDate(1, 0, 0)
	_, err := Query(
		CreatedBetween(future, now),
		ClosedBetween(future, now),
	)
	assert.Error(t, err)
}

// Examples

func ExampleQuery() {
	end := time.Date(2025, 01, 13, 10, 0, 0, 0, time.Local)
	start := end.AddDate(0, 0, -2)
	query, err := Query(
		Organization("testOrg"),
		Repository("testRepo"),
		ClosedBetween(start, end),
	)
	if err != nil {
		// something went wrong
	}
	fmt.Print(query)
	// Output:
	// org:testOrg repo:testRepo closed:2025-01-11..2025-01-13
}

// Benchmarks

var res GithubQuery

func BenchmarkRepository(b *testing.B) {
	var r GithubQuery
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		r = Repository(testOrgRepo)
	}
	res = r
}
