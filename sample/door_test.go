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

		It("Opens door when correct password is given", func() {
			sample.Button(genFsm, 'p')
			sample.Button(genFsm, 'a')
			sample.Button(genFsm, 's')
			sample.Button(genFsm, 's')
			genFsm.Wait()
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Open")))
		})

		It("Opens door and locks after timeout expires", func() {
			sample.Button(genFsm, 'p')
			sample.Button(genFsm, 'a')
			sample.Button(genFsm, 's')
			sample.Button(genFsm, 's')
			genFsm.Wait()
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Open")))
			time.Sleep(2 * time.Second)
			Expect(genFsm.GetCurrentState()).Should(Equal(go_gen_fsm.State("Locked")))
		})

	})

})
