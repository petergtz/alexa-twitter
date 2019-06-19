package verbalizer_test

import (
	"testing"

	"github.com/petergtz/pegomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVerbalizer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Verbalizer Suite")
	pegomock.RegisterMockFailHandler(Fail)
}
