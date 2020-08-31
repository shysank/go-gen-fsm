package go_gen_fsm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"time"
)

var _ = Describe("GenFsm", func() {

	const (
		State1 = State("State1")
		State2 = State("State2")
		State3 = State("State3")
		Event1 = Event("Event1")
		Event2 = Event("Event2")
		Event3 = Event("Event3")
	)

	var (
		genFsm  *GenFSM
		testFsm *testFsmMock
	)

	Context("Start", func() {
		BeforeEach(func() {
			testFsm = &testFsmMock{}
			testFsm.mock.On("Init", "arg1", "arg2").Return(State1)
		})

		It("Sets Initial state and Registers Handlers", func() {
			genFsm = Start(testFsm, "arg1", "arg2")
			Expect(genFsm.currentState).Should(Equal(State1))
			Expect(len(genFsm.handlers)).Should(Equal(2))

			Expect(len(genFsm.handlers[State1])).Should(Equal(2))
			Expect(len(genFsm.handlers[State2])).Should(Equal(2))

			Expect(genFsm.handlers[State1][0].event).Should(Equal(Event1))
			Expect(genFsm.handlers[State1][1].event).Should(Equal(Event2))
			Expect(genFsm.handlers[State2][0].event).Should(Equal(Event1))
			Expect(genFsm.handlers[State2][1].event).Should(Equal(Event3))

			genFsm.Stop()
		})
	})

	Context("SendEvent", func() {
		BeforeEach(func() {
			testFsm = &testFsmMock{}
			testFsm.mock.On("Init", "arg1", "arg2").Return(State1)
			genFsm = Start(testFsm, "arg1", "arg2")
			testFsm.mock.On("State1_Event1", 100, true, "SomeString").Return(State2)
			testFsm.mock.On("State1_Event1", 100, true, "GoToState3").Return(State3)
		})

		AfterEach(func() {
			genFsm.Stop()
		})

		Context("Success", func() {

			It("Dispatches to right handler and updates state", func() {
				genFsm.SendEvent(Event1, 100, true, "SomeString")
				time.Sleep(100 * time.Microsecond)
				Expect(genFsm.currentState).Should(Equal(State2))
			})

			It("Updates to a different state for different args", func() {
				genFsm.SendEvent(Event1, 100, true, "GoToState3")
				time.Sleep(100 * time.Microsecond)
				Expect(genFsm.currentState).Should(Equal(State3))
			})
		})

		Context("Fail", func() {

			It("Incorrect no of args", func() {
				genFsm.SendEvent(Event1, 100)
				err := <-genFsm.errorChannel
				Expect(err.Error()).Should(Equal("reflect: Call with too few input arguments"))
				Expect(genFsm.currentState).Should(Equal(State1))
			})

			It("Invalid event", func() {
				genFsm.SendEvent("Invalid Event", 100, true, "SomeString")
				time.Sleep(100 * time.Microsecond)
				Expect(genFsm.currentState).Should(Equal(State1))
			})

			It("Invalid state", func() {
				genFsm.currentState = "Invalid State"
				genFsm.SendEvent(Event1, 100, true, "SomeString")
				time.Sleep(100 * time.Microsecond)
				Expect(genFsm.currentState).Should(Equal(State("Invalid State")))
			})
		})

	})

	Context("SendSync", func() {
		BeforeEach(func() {
			testFsm = &testFsmMock{}
			testFsm.mock.On("Init", "arg1", "arg2").Return(State1)
			genFsm = Start(testFsm, "arg1", "arg2")
			testFsm.mock.On("State1_Event1", 100, true, "SomeString").Return(State2)
			testFsm.mock.On("State1_Event1", 100, true, "GoToState3").Return(State3)
		})

		It("Sends no op request and gets response", func() {
			genFsm.SendEvent(Event1, 100, true, "SomeString")
			resp := genFsm.SendSyncReq(NOOP)
			Expect(genFsm.currentState).Should(Equal(State2))
			Expect(resp).Should(Equal(NOOP))
			genFsm.Stop()
		})

		It("Gets 404 if there is no handler registered for the request", func() {
			genFsm.SendEvent(Event1, 100, true, "SomeString")
			resp := genFsm.SendSyncReq("Invalid_Req")
			Expect(genFsm.currentState).Should(Equal(State2))
			Expect(resp).Should(Equal(404))
			genFsm.Stop()
		})

		It("Shuts down gen fsm go routine ", func() {
			genFsm.SendEvent(Event1, 100, true, "SomeString")
			resp := genFsm.SendSyncReq(STOP)
			Expect(genFsm.currentState).Should(Equal(State2))
			Expect(genFsm.shutdown).Should(Equal(true))
			Expect(resp).Should(Equal(STOP))
			Expect(func() { genFsm.SendEvent(Event1) }).To(Panic())
		})

	})

	Context("Start to finish", func() {
		BeforeEach(func() {
			testFsm = &testFsmMock{}
			testFsm.mock.On("Init", "arg1", "arg2").Return(State1)
			genFsm = Start(testFsm, "arg1", "arg2")
			testFsm.mock.On("State1_Event1", 100, true, "SomeString").Return(State2)
			testFsm.mock.On("State1_Event1", 100, true, "GoToState3").Return(State3)
			testFsm.mock.On("State2_Event1", 101, false, "SomeString2").Return(State1)
			testFsm.mock.On("State1_Event2", 102, false, "SomeString3").Return(State2)
			testFsm.mock.On("State2_Event3", 999, true, "EndIt").Return(State("End"))
		})

		It("Init -> Dispatch Event -> Next State -> Dispatch Event..", func() {
			genFsm.SendEvent(Event1, 100, true, "SomeString")
			time.Sleep(100 * time.Microsecond)
			Expect(genFsm.currentState).Should(Equal(State2))

			genFsm.SendEvent(Event1, 101, false, "SomeString2")
			time.Sleep(100 * time.Microsecond)
			Expect(genFsm.currentState).Should(Equal(State1))

			genFsm.SendEvent(Event2, 102, false, "SomeString3")
			time.Sleep(100 * time.Microsecond)
			Expect(genFsm.currentState).Should(Equal(State2))

			genFsm.SendEvent(Event3, 999, true, "EndIt")
			time.Sleep(100 * time.Microsecond)
			Expect(genFsm.currentState).Should(Equal(State("End")))
		})

		AfterEach(func() {
			genFsm.Stop()
		})
	})
})

type testFsmMock struct {
	mock mock.Mock
}

// Generate this code later
type testFsm interface {
	FSM
	State1_Event1(arg1 int, arg2 bool, arg3 string) State
	State1_Event2(arg1 int, arg2 bool, arg3 string) State
	State2_Event1(arg1 int, arg2 bool, arg3 string) State
	State2_Event3(arg1 int, arg2 bool, arg3 string) State
}

func (t *testFsmMock) Init(args ...interface{}) State {
	a := t.mock.Called(args...)
	return a.Get(0).(State)
}

func (t *testFsmMock) State1_Event1(arg1 int, arg2 bool, arg3 string) State {
	args := t.mock.Called(arg1, arg2, arg3)
	return args.Get(0).(State)
}

func (t *testFsmMock) State1_Event2(arg1 int, arg2 bool, arg3 string) State {
	args := t.mock.Called(arg1, arg2, arg3)
	return args.Get(0).(State)
}

func (t *testFsmMock) State2_Event1(arg1 int, arg2 bool, arg3 string) State {
	args := t.mock.Called(arg1, arg2, arg3)
	return args.Get(0).(State)
}

func (t *testFsmMock) State2_Event3(arg1 int, arg2 bool, arg3 string) State {
	args := t.mock.Called(arg1, arg2, arg3)
	return args.Get(0).(State)
}
