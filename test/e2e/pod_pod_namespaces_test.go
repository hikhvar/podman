package integration

import (
	"fmt"
	"os"

	. "github.com/containers/podman/v3/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Podman pod create", func() {
	var (
		tempdir    string
		err        error
		podmanTest *PodmanTestIntegration
	)

	BeforeEach(func() {
		tempdir, err = CreateTempDirInTempDir()
		if err != nil {
			os.Exit(1)
		}
		podmanTest = PodmanTestCreate(tempdir)
		podmanTest.Setup()
		podmanTest.SeedImages()
	})

	AfterEach(func() {
		podmanTest.Cleanup()
		f := CurrentGinkgoTestDescription()
		processTestResult(f)

	})

	It("podman pod container share Namespaces", func() {
		session := podmanTest.Podman([]string{"pod", "create"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		podID := session.OutputToString()

		session = podmanTest.Podman([]string{"pod", "start", podID})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		session = podmanTest.Podman([]string{"run", "--pod", podID, "-d", ALPINE, "top"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		check := podmanTest.Podman([]string{"ps", "-a", "--ns", "--format", "{{.Namespaces.IPC}} {{.Namespaces.UTS}} {{.Namespaces.NET}}"})
		check.WaitWithDefaultTimeout()
		Expect(check).Should(Exit(0))
		outputArray := check.OutputToStringArray()
		Expect(outputArray).To(HaveLen(2))

		NAMESPACE1 := outputArray[0]
		fmt.Println("NAMESPACE1:", NAMESPACE1)
		NAMESPACE2 := outputArray[1]
		fmt.Println("NAMESPACE2:", NAMESPACE2)
		Expect(NAMESPACE1).To(Equal(NAMESPACE2))
	})

	It("podman pod container share ipc && /dev/shm ", func() {
		session := podmanTest.Podman([]string{"pod", "create"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		podID := session.OutputToString()

		session = podmanTest.Podman([]string{"pod", "start", podID})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		session = podmanTest.Podman([]string{"run", "--rm", "--pod", podID, ALPINE, "touch", "/dev/shm/test"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		session = podmanTest.Podman([]string{"run", "--rm", "--pod", podID, ALPINE, "ls", "/dev/shm/test"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
	})

	It("podman pod container dontshare PIDNS", func() {
		session := podmanTest.Podman([]string{"pod", "create"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))
		podID := session.OutputToString()

		session = podmanTest.Podman([]string{"pod", "start", podID})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		session = podmanTest.Podman([]string{"run", "--pod", podID, "-d", ALPINE, "top"})
		session.WaitWithDefaultTimeout()
		Expect(session).Should(Exit(0))

		check := podmanTest.Podman([]string{"ps", "-a", "--ns", "--format", "{{.Namespaces.PIDNS}}"})
		check.WaitWithDefaultTimeout()
		Expect(check).Should(Exit(0))
		outputArray := check.OutputToStringArray()
		Expect(outputArray).To(HaveLen(2))

		NAMESPACE1 := outputArray[0]
		fmt.Println("NAMESPACE1:", NAMESPACE1)
		NAMESPACE2 := outputArray[1]
		fmt.Println("NAMESPACE2:", NAMESPACE2)
		Expect(NAMESPACE1).To(Not(Equal(NAMESPACE2)))
	})

})
