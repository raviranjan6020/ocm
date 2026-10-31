package e2e

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Fork-only test branch: fails on purpose to check the sharded summary. Not for the PR.
var _ = ginkgo.Describe("Injected failure", func() {
	ginkgo.It("fails on purpose in the addon shard", func() {
		gomega.Expect("addon shard").To(gomega.Equal("injected failure"))
	})
})
