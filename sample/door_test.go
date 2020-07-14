package sample_test

import (
	go_gen_fsm "github.com/cynic89/go-gen-fsm"
	"github.com/cynic89/go-gen-fsm/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Door", func() {

	var (
		subject *sample.Door
		genFsm  *go_gen_fsm.GenFSM
	)

	Context("Door with password 'pass'", func() {
		BeforeEach(func() {
			subject = new(sample.Door)
			genFsm = go_gen_fsm.Start(subject, "pass")
		})

		AfterEach(func() {
			genFsm.Stop()
		})

		It("Opens door when correct password is given", func() {
			enterCode("pass", genFsm)
			genFsm.Wait()
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Open")))
		})

		It("Does not open door when incorrect password is given", func() {
			enterCode("rong", genFsm)
			genFsm.Wait()
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Locked")))
		})

		It("Opens door and locks after timeout expires", func() {
			enterCode("pass", genFsm)
			genFsm.Wait()
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Open")))
			time.Sleep(sample.LockTimeout + 1*time.Second)
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Locked")))
		})
	})

})

func enterCode(pass string, genFsm *go_gen_fsm.GenFSM) {
	for _, c := range pass {
		sample.Button(genFsm, c)
	}
}
