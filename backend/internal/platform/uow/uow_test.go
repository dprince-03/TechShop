package uow

import (
	"testing"

	"github.com/riverqueue/river"
)

type testArgs struct{ N int }

func (testArgs) Kind() string { return "test" }

func TestDedupeJobsDropsExactDuplicates(t *testing.T) {
	in := []river.InsertManyParams{{Args: testArgs{1}}, {Args: testArgs{2}}, {Args: testArgs{1}}}
	out := dedupeJobs(in)
	if len(out) != 2 || out[0].Args.(testArgs).N != 1 || out[1].Args.(testArgs).N != 2 {
		t.Fatalf("dedupe = %+v", out)
	}
	if len(in) != 3 {
		t.Fatal("input must not be modified")
	}
}
