package go_gen_fsm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenFsm", func() {

	var (
		subject *GenFSM
	)

	Context("Start", func() {
		BeforeEach(func() {
			testFsm := &testFsm{}
			subject = Start(testFsm)
		})

		It("Starts gen_fsm go routine", func() {
			Expect(subject.currentState).Should(Equal(State("InitialState")))
			Expect(len(subject.handlers)).Should(Equal(1))
		})
	})

	It("Sample test", func() {
		Expect(1).Should(Equal(1))
	})
})

type testFsm struct {
}

func (t *testFsm) Init(args ...interface{}) State {
	return "InitialState"
}

func (t *testFsm) InitialState_SomeEvent() State {
	return "State2"
}
