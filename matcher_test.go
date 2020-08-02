package go_gen_fsm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DelimiterMatcher", func() {
	var (
		subject NameMatcher
	)
	Context("Match", func() {

		BeforeEach(func() {
			subject = DelimiterMatcher{"_"}
		})

		DescribeTable("Matches with pattern StateName_EventName", func(name, stateExp, eventExp string) {
			result, state, event := subject.Matches(name)
			Expect(result).Should(Equal(true))
			Expect(state).Should(Equal(State(stateExp)))
			Expect(event).Should(Equal(Event(eventExp)))

		}, Entry("Simple", "State_Event", "State", "Event"),
			Entry("MultiWord", "StateWithMoreWords_EventWithMoreWords", "StateWithMoreWords", "EventWithMoreWords"),
			Entry("EventWithLowerCaseLetter", "State_eventWithLowerCase", "State", "eventWithLowerCase"))
	})

	Context("Does Not Match", func() {

		BeforeEach(func() {
			subject = DelimiterMatcher{"_"}
		})

		DescribeTable("Does not match with pattern other than StateName_EventName", func(name string) {
			result, _, _ := subject.Matches(name)
			Expect(result).Should(Equal(false))

		}, Entry("MultipleUnderscores", "State_Ev_ent"),
			Entry("OtherSpecialCharacters", "State#Event"),
			Entry("LowercaseState", "state_Event"))

	})
})
