package save_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/pkg/virtctl/save"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Saving", func() {

	const vmName = "testvm"
	const restoreFile = "test-restore"
	var vmInterface *kubecli.MockVirtualMachineInterface
	var vmiInterface *kubecli.MockVirtualMachineInstanceInterface
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		kubecli.GetKubevirtClientFromClientConfig = kubecli.GetMockKubevirtClientFromClientConfig
		kubecli.MockKubevirtClientInstance = kubecli.NewMockKubevirtClient(ctrl)
		vmInterface = kubecli.NewMockVirtualMachineInterface(ctrl)
		vmiInterface = kubecli.NewMockVirtualMachineInstanceInterface(ctrl)
	})

	Context("With missing input parameters", func() {
		It("should fail", func() {
			cmd := tests.NewRepeatableVirtctlCommand(save.COMMAND_SAVE)
			err := cmd()
			Expect(err).NotTo(BeNil())
		})
	})

	It("should save VMI", func() {
		vmi := v1.NewMinimalVMI(vmName)

		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)
		vmiInterface.EXPECT().Save(vmi.Name).Return(nil).Times(1)

		cmd := tests.NewVirtctlCommand(save.COMMAND_SAVE, "vmi", vmName)
		Expect(cmd.Execute()).To(BeNil())
	})

	It("should restore VMI", func() {
		vmi := v1.NewMinimalVMI(vmName)

		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)
		vmiInterface.EXPECT().Restore(vmi.Name, restoreFile).Return(nil).Times(1)

		cmd := tests.NewVirtctlCommand(save.COMMAND_RESTORE, "vmi", vmName)
		Expect(cmd.Execute()).To(BeNil())
	})

	It("should save VM", func() {
		vmi := v1.NewMinimalVMI(vmName)
		vm := kubecli.NewMinimalVM(vmName)
		vm.Spec.Template = &v1.VirtualMachineInstanceTemplateSpec{
			Spec: vmi.Spec,
		}

		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)

		vmInterface.EXPECT().Get(vm.Name, &k8smetav1.GetOptions{}).Return(vm, nil).Times(1)
		vmiInterface.EXPECT().Save(vm.Name).Return(nil).Times(1)

		cmd := tests.NewVirtctlCommand(save.COMMAND_SAVE, "vm", vmName)
		Expect(cmd.Execute()).To(BeNil())
	})

	It("should restore VM", func() {
		vmi := v1.NewMinimalVMI(vmName)
		vm := kubecli.NewMinimalVM(vmName)
		vm.Spec.Template = &v1.VirtualMachineInstanceTemplateSpec{
			Spec: vmi.Spec,
		}

		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachine(k8smetav1.NamespaceDefault).Return(vmInterface).Times(1)
		kubecli.MockKubevirtClientInstance.EXPECT().VirtualMachineInstance(k8smetav1.NamespaceDefault).Return(vmiInterface).Times(1)

		vmInterface.EXPECT().Get(vm.Name, &k8smetav1.GetOptions{}).Return(vm, nil).Times(1)
		vmiInterface.EXPECT().Restore(vm.Name, restoreFile).Return(nil).Times(1)

		cmd := tests.NewVirtctlCommand(save.COMMAND_RESTORE, "vm", vmName)
		Expect(cmd.Execute()).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})
})
