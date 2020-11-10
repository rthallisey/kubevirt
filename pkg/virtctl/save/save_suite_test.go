package save_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"kubevirt.io/client-go/log"
)

func TestSaveRestore(t *testing.T) {
	log.Log.SetIOWriter(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Saving Suite")
}
